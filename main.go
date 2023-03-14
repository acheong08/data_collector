package main

import (
	"os"

	"github.com/acheong08/data_collector/internal/server"
	gin "github.com/gin-gonic/gin"
)

func adminMiddleware(c *gin.Context) {
	// Get the auth token from the request header
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		c.JSON(401, gin.H{"error": "Missing Authorization header"})
		c.Abort()
		return
	}
	// Check if the token is valid
	if authToken != os.Getenv("AUTH_TOKEN") {
		c.JSON(401, gin.H{"error": "Invalid Authorization token"})
		c.Abort()
		return
	}
	c.Next()
}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/analytics/message", server.Message)
	r.POST("/exit", adminMiddleware, server.Exit)
	r.POST("/reset", adminMiddleware, server.Reset)
	r.Run() // listen and serve on
}
