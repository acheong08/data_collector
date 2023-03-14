package main

import (
	"context"
	"fmt"
	"log"
	"os"

	typings "github.com/acheong08/data_collector/internal/typings"
	gin "github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	pgx "github.com/jackc/pgx/v5"
)

var db *pgx.Conn
var err error

func init() {
	// Source .env file
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	rows := db.QueryRow(context.Background(), "select version()")
	if err != nil {
		log.Fatal(err)
	}
	var version string
	err = rows.Scan(&version)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("DB version=%s\n", version)
	// Create the conversations table if it doesn't exist
	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS conversations ( id TEXT PRIMARY KEY, "user" VARCHAR(255), messages JSONB )`)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/collect", Collect)
	r.POST("/exit", func(c *gin.Context) {
		// Check authentication
		if c.Request.Header.Get("Authorization") != os.Getenv("AUTH") {
			return
		}
		// Close the database connection
		err = db.Close(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		// Stop the program
		os.Exit(0)
	})
	r.Run() // listen and serve on
}

// Collect is a handler which stores the conversation in the database
func Collect(c *gin.Context) {
	var conversation typings.Conversation
	err := c.BindJSON(&conversation)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// Store the conversation in the database
	_, err = db.Exec(context.Background(), `INSERT INTO conversations (id, "user", messages) VALUES ($1, $2, $3)`, conversation.Id, conversation.User, conversation.Messages)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}
