package main

import (
	"embed"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/filesystem"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	"github.com/0verbyte/twc-gen3/internal/storage"
	"github.com/0verbyte/twc-gen3/pkg/twc"
)

type APIv1 struct {
	twc     *twc.TWC
	storage *storage.DB
}

func NewAPIv1() (*APIv1, error) {
	db, err := storage.New()
	if err != nil {
		return nil, err
	}

	if err := db.CreateTables(); err != nil {
		log.WithError(err).Fatal("Failed to create database tables")
	}

	if ip, err := db.GetTWCIP(); err == nil {
		twc, err := twc.New(ip)
		if err == nil {
			return &APIv1{
				twc:     twc,
				storage: db,
			}, nil
		} else {
			log.WithError(err).Warnf("Failed to create twc %s", ip)
		}
	} else {
		log.WithError(err).Warn("Failed to lookup TWC in storage")
	}

	ip := os.Getenv("TWC_IP")
	if ip != "" {
		twc, err := twc.New(ip)
		if err == nil {
			log.Debugf("Using wall connector from TWC_IP=%s env var", twc.IP())
			return &APIv1{
				twc:     twc,
				storage: db,
			}, nil
		}
		log.WithError(err).Warnf("Failed to use TWC_IP=%s", ip)
	}

	twc, err := twc.Find()
	if err != nil {
		log.WithError(err).Fatalln("Failed to find twc on network")
	}

	if err := db.SaveTWCIP(twc.IP()); err != nil {
		log.WithError(err).Error("Failed to save twc ip to database")
	}

	return &APIv1{
		twc:     twc,
		storage: db,
	}, nil
}

func (api *APIv1) Vitals(c fiber.Ctx) error {
	vitals, err := api.twc.GetVitals()
	if err != nil {
		log.WithError(err).Error("Failed to get Tesla Wall Connector vitals")
		c.Status(fiber.StatusInternalServerError).SendString("Failed to get vitals")
	}

	return c.Status(fiber.StatusOK).JSON(vitals)
}

func (api *APIv1) WifiStatus(c fiber.Ctx) error {
	status, err := api.twc.GetWifiStatus()
	if err != nil {
		log.WithError(err).Error("Failed to get Tesla Wall Connector wifi status")
		c.Status(fiber.StatusInternalServerError).SendString("Failed to get wifi status")
	}

	return c.Status(fiber.StatusOK).JSON(status)
}

func (api *APIv1) Lifetime(c fiber.Ctx) error {
	stats, err := api.twc.GetLifetimeStats()
	if err != nil {
		log.WithError(err).Error("Failed to get Tesla Wall Connector lifetime status")
		c.Status(fiber.StatusInternalServerError).SendString("Failed to get lifetime stats")
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}

func (api *APIv1) Info(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(map[string]string{
		"ip": api.twc.IP(),
	})
}

func (api *APIv1) TWCConnectedMiddleware(c fiber.Ctx) error {
	// Skip because we are trying to find the twc connector when calling this path
	if strings.HasSuffix(c.Path(), "/find") {
		return c.Next()
	}

	if api.twc == nil {
		return c.Status(fiber.StatusOK).JSON(map[string]string{
			"error": "not connected to Tesla Wall Connector",
		})
	}
	return c.Next()
}

func (api *APIv1) Find(c fiber.Ctx) error {
	if api.twc != nil {
		return c.Status(fiber.StatusOK).JSON(map[string]string{
			"ip":     api.twc.IP(),
			"status": "connected",
		})
	}

	twc, err := twc.Find()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
			"error": err.Error(),
		})
	}

	api.twc = twc

	if err := api.storage.SaveTWCIP(twc.IP()); err != nil {
		log.WithError(err).Warn("Failed to save twc to storage")
	}

	return c.Status(fiber.StatusOK).JSON(map[string]string{
		"status": "connected",
		"ip":     twc.IP(),
	})
}

var (
	//go:embed web/build/index.html
	f embed.FS

	//go:embed web/build/static
	staticDir embed.FS
)

func main() {
	log.SetLevel(log.DebugLevel)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		log.Infof("[%s] %s (%d) %s - %s", c.Context().Time().Format(time.RFC822), c.Method(), c.Response().StatusCode(), c.Protocol(), c.Path())
		return c.Next()
	})

	_, err := f.ReadFile("web/build/index.html")
	if err != nil {
		log.WithError(err).Fatal("Failed to create static embed for index.html")
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:  f,
		Index: "web/build/index.html",
	}))
	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       staticDir,
		PathPrefix: "web/build/static",
		Browse:     true,
	}))

	api := app.Group("/api")
	v1 := api.Group("/v1")
	apiv1, err := NewAPIv1()
	if err != nil {
		log.WithError(err).Fatal("Error creating API v1")
	}

	v1.Use(apiv1.TWCConnectedMiddleware)

	v1.Get("/wifi_status", apiv1.WifiStatus)
	v1.Get("/vitals", apiv1.Vitals)
	v1.Get("/lifetime", apiv1.Lifetime)
	v1.Get("/info", apiv1.Info)
	v1.Get("/find", apiv1.Find)

	log.Fatal(app.Listen(":8080"))
}
