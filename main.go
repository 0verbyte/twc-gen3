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
	storage, err := storage.New()
	if err != nil {
		return nil, err
	}

	if err := storage.Init(); err != nil {
		log.WithError(err).Fatal("Failed to init storage")
	}

	if ip, err := storage.GetTWCIP(); err == nil {
		twc, err := twc.New(ip)
		if err == nil {
			return &APIv1{
				twc:     twc,
				storage: storage,
			}, nil
		}
		log.WithError(err).Warnf("Failed to create TWC from %s found in storage", ip)
	} else {
		log.Debug("No record in storage for TWC")
	}

	ip := os.Getenv("TWC_IP")
	if ip != "" {
		twc, err := twc.New(ip)
		if err == nil {
			if _, err := twc.GetVitals(); err == nil {
				if err := storage.SaveTWCIP(twc.IP()); err != nil {
					log.WithError(err).Warnf("Failed to save TWC (%s) to storage", twc.IP())
				}

				return &APIv1{
					twc:     twc,
					storage: storage,
				}, nil
			} else {
				log.WithError(err).Warnf("TWC get vitals failed for %s", twc.IP())
			}
		}
		log.WithError(err).Warnf("Failed to create TWC from %s found in TWC_IP environment variable", ip)
	} else {
		log.Debug("TWC_IP environment variable is not set or is empty")
	}

	twc, err := twc.Find()
	if err != nil {
		log.WithError(err).Fatalln("Failed to find TWC")
	}

	if err := storage.SaveTWCIP(twc.IP()); err != nil {
		log.WithError(err).Fatalln("Failed to save twc ip to database")
	}

	return &APIv1{
		twc:     twc,
		storage: storage,
	}, nil
}

func (api *APIv1) Vitals(c fiber.Ctx) error {
	vitals, err := api.twc.GetVitals()
	if err != nil {
		log.WithError(err).Error("Failed to get Tesla Wall Connector vitals")
		c.Status(fiber.StatusInternalServerError).SendString("Failed to get vitals")
	}

	if err := api.storage.RecordVital(api.twc.IP(), vitals); err != nil {
		log.WithError(err).Warnf("Failed to record vital")
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

func (api *APIv1) Query(c fiber.Ctx) error {
	queryDuration := c.Query("duration")
	durationString := "15m"
	if queryDuration != "" {
		durationString = queryDuration
	}

	duration, err := time.ParseDuration(durationString)
	if err != nil {
		log.WithError(err).Errorf("Failed to parse duration string %s", durationString)
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
			"error": "invalid duration " + queryDuration,
		})
	}

	startTime := time.Now().Add(-(duration))
	vitals, err := api.storage.QueryVitals(startTime)
	if err != nil {
		log.WithError(err).Errorf("Failed to query vitals for range %s", startTime)
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
			"error": "query vitals failed",
		})
	}

	return c.Status(fiber.StatusOK).JSON(vitals)
}

var (
	//go:embed web/build/index.html
	f embed.FS

	//go:embed web/build/static
	staticDir embed.FS
)

func pollVitals(twc *twc.TWC, storage *storage.DB) {
	log.Debugf("poll vitals started for twc %s", twc.IP())
	const duration = time.Second * 1
	for {
		log.Debugf("Checking twc vitals in %s", duration.String())
		time.Sleep(duration)

		log.Debugf("Get vitals for %s", twc.IP())
		vital, err := twc.GetVitals()
		if err != nil {
			log.WithError(err).Error("Failed to get vitals")
			continue
		}
		log.Debugf("Got vitals for %s", twc.IP())
		if err := storage.RecordVital(twc.IP(), vital); err != nil {
			log.WithError(err).Error("Failed saving vital record to storage")
		}

		log.Debugf("Saved vitals for %s to storage", twc.IP())
	}
}

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

	log.Infof("TWC connected: %s", apiv1.twc.IP())

	go pollVitals(apiv1.twc, apiv1.storage)

	v1.Use(apiv1.TWCConnectedMiddleware)

	v1.Get("/wifi_status", apiv1.WifiStatus)
	v1.Get("/vitals", apiv1.Vitals)
	v1.Get("/lifetime", apiv1.Lifetime)
	v1.Get("/info", apiv1.Info)
	v1.Get("/find", apiv1.Find)
	v1.Get("/query", apiv1.Query)

	log.Fatal(app.Listen(":8080"))
}
