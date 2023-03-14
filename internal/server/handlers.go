package server

import (
	"context"
	"log"
	"os"

	"github.com/acheong08/data_collector/internal/typings"
	"github.com/gin-gonic/gin"
	pgx "github.com/jackc/pgx/v5"
)

var db *pgx.Conn
var err error

func init() {
	db, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
}

// h_message is a handler which stores the conversation in the database
func Message(c *gin.Context) {
	if db.IsClosed() {
		db, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}
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
	// Append the new message to the messages array (postgres 14 compatible)
	_, err = db.Exec(
		context.Background(), `
			INSERT INTO conversations (id, "user", messages)
			VALUES ($1, $2, ARRAY[to_jsonb($3::jsonb)])
			ON CONFLICT (id) DO UPDATE
			SET messages = conversations.messages || ARRAY[to_jsonb($3::jsonb)]
			WHERE conversations.id = $1
		`,
		msg_instance.Id, msg_instance.User, msg_instance.Message)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "success"})
}

func Reset(c *gin.Context) {
	// Delete the conversations table if it exists
	_, err := db.Exec(context.Background(), `DROP TABLE IF EXISTS conversations`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	// Create the conversations table if it doesn't exist
	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS conversations ( id TEXT PRIMARY KEY NOT NULL, "user" VARCHAR(32) NOT NULL, messages JSONB[] NOT NULL )`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})

}

func Exit(c *gin.Context) {
	// Close the database connection
	err := db.Close(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
	// Stop the program
	os.Exit(0)
}
