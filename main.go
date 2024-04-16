package main

import (
	"database/sql"
	"embed"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/filesystem"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	"github.com/0verbyte/twc-gen3/pkg/twc"
)

type APIv1 struct {
	twc     *twc.TWC
	storage *sql.DB
}

func NewAPIv1() (*APIv1, error) {
	db, err := sql.Open("sqlite3", "./twc_gen3.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("create table if not exists twc (ip text primary key)")
	if err != nil {
		return nil, err
	}

	row := db.QueryRow("SELECT ip from twc")
	var ip string
	if err := row.Scan(&ip); err == nil && len(ip) > 0 {
		twc, err := twc.New(ip)
		if err == nil {
			log.Debugf("Using wall connector from storage db %s", twc.IP())
			return &APIv1{
				twc:     twc,
				storage: db,
			}, nil
		}
		log.WithError(err).Warnf("Failed to load twc %s", ip)
	}

	ip = os.Getenv("TWC_IP")
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

	tx, err := api.storage.Prepare("INSERT into twc (ip) values(?)")
	if err != nil {
		log.WithError(err).Error("Failed to prepare twc ip database statement")
	}
	if _, err := tx.Exec(twc.IP()); err != nil {
		log.WithError(err).Error("Failed to execute twc ip database statement")
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
