package main // import "github.com/marksteve/radioslack"

import (
	"log"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/marksteve/radioslack/radioslack"
)

func main() {
	e := echo.New()
	e.Use(mw.Logger())
	e.Use(mw.Recover())

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
