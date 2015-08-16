package main // import "github.com/marksteve/radioslack"

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/marksteve/radioslack/radioslack"
)

type Config struct {
	SlackClientId     string
	SlackClientSecret string
}

func ConfigContext() echo.HandlerFunc {
	var config Config
	err := envconfig.Process("rs", &config)
	if err != nil {
		log.Fatal(err.Error())
	}
	return func(c *echo.Context) error {
		c.Set("SlackClientId", config.SlackClientId)
		c.Set("SlackClientSecret", config.SlackClientSecret)
		return nil
	}
}

func main() {
	e := echo.New()
	e.Use(mw.Logger())
	e.Use(mw.Recover())
	e.Use(ConfigContext())

	e.Static("/", "static")
	e.Get("/login", radioslack.Login)
	e.Get("/oauth", radioslack.OAuth)
	e.Get("/me", radioslack.Me)
	e.Get("/logout", radioslack.Logout)
	e.WebSocket("/queue/:teamId/:ch", radioslack.Queue)

	go radioslack.Start()

	addr := ":8000"
	log.Print("RADIOSLACK")
	log.Printf("Listening on %s", addr)
	e.Run(addr)
}
