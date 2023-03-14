package server

import (
	"context"
	"log"
	"os"

	"github.com/acheong08/data_collector/internal/typings"
	"github.com/gin-gonic/gin"
	pgx "github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
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
}

// h_message is a handler which stores the conversation in the database
func Message(c *gin.Context) {
	var msg_instance typings.MessageInstance
	err := c.BindJSON(&msg_instance)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}
	// If any of the fields are empty, return an error
	if msg_instance.Id == "" || msg_instance.User == "" || msg_instance.Message == (typings.Message{}) {
		c.JSON(400, gin.H{"error": "Missing fields"})
		return
	}

	// INSERT if the conversation doesn't exist, UPDATE and append new message if it does
	_, err = db.Exec(context.Background(), `INSERT INTO conversations (id, "user", messages) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET messages = array_append(conversations.messages, $3)`, msg_instance.Id, msg_instance.User, msg_instance.Message)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}

	c.JSON(200, gin.H{"message": "success"})
}

func Reset(c *gin.Context) {
	// Delete the conversations table if it exists
	_, err := db.Exec(context.Background(), `DROP TABLE IF EXISTS conversations`)
	// Create the conversations table if it doesn't exist
	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS conversations ( id TEXT PRIMARY KEY NOT NULL, "user" VARCHAR(255) NOT NULL, messages JSONB[] NOT NULL )`)
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(200, gin.H{"message": "success"})

}

func Exit(c *gin.Context) {
	// Close the database connection
	err := db.Close(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// Stop the program
	os.Exit(0)
}
