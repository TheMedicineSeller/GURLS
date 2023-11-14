package main

import (
	"github.com/TheMedicineSeller/GURLS/config"
	"github.com/TheMedicineSeller/GURLS/routes"
	"github.com/gin-gonic/gin"
)

// Define endpoints for app and attach corresponding functions to the endpoints with the right method

func main() {
	app := gin.Default()
	app.GET("/:url", routes.ResolveURL)
	app.POST("/api", routes.ShortenURL)

	app.Run(config.PORT)
}
