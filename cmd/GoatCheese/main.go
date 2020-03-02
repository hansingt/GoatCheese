package main

import (
	"flag"
	"github.com/hansingt/GoatCheese/internal/datastore"
	"github.com/hansingt/GoatCheese/internal/web"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	configurationFile := flag.String(
		"config",
		"config.yaml",
		"Path to the YAML file to read the configuration from")
	templatesPath := flag.String(
		"templates",
		"./templates",
		"Path to the directory containing the HTML templates to serve (default: ./templates)")
	flag.Parse()
	db, err := datastore.New(*configurationFile)
	if err != nil {
		panic(err)
	}

	server := echo.New()
	// Setup the Middleware
	server.Use(middleware.Logger())
	server.Use(middleware.Recover())
	// Setup the routes
	if err := web.SetupEchoServer(server, db, *templatesPath); err != nil {
		panic(err)
	}
	// Start the server
	server.Logger.Fatal(server.Start(":8080"))
}
