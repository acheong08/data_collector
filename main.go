package main

import (
	"context"
	"log"
	"os"
	"regexp"

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
}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/analytics/new_convo", StartConversation)
	r.POST("/analytics/add_message", AddMessage)
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
	r.POST("/reset", func(c *gin.Context) {
		// Delete the conversations table if it exists
		_, err = db.Exec(context.Background(), `DROP TABLE IF EXISTS conversations`)
		// Create the conversations table if it doesn't exist
		_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS conversations ( id TEXT PRIMARY KEY NOT NULL, "user" VARCHAR(255) NOT NULL, messages JSONB[] NOT NULL )`)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(200, gin.H{"message": "success"})

	})
	r.Run() // listen and serve on
}

// StartConversation is a handler which stores the conversation in the database
func StartConversation(c *gin.Context) {
	var conversation typings.Conversation
	err := c.BindJSON(&conversation)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}
	// If any of the fields are empty, return an error
	if conversation.Id == "" || conversation.User == "" || len(conversation.Messages) == 0 {
		c.JSON(400, gin.H{"error": "Missing fields"})
		return
	}

	// Store the conversation in the database
	_, err = db.Exec(context.Background(), `INSERT INTO conversations (id, "user", messages) VALUES ($1, $2, $3)`, conversation.Id, conversation.User, conversation.Messages)
	if err != nil {
		err_code := 500
		err_string := err.Error()
		// Regex check for duplicate key value violates unique constraint
		if regexp.MustCompile(`duplicate key value violates unique constraint`).MatchString(err.Error()) {
			err_code = 409
			err_string = "Conversation already exists"
		}
		c.JSON(err_code, gin.H{"error": err_string})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}

type StandaloneMessage struct {
	Message typings.Message `json:"message"`
	ConvoId string          `json:"convo_id"`
}

func AddMessage(c *gin.Context) {
	var standaloneMessage StandaloneMessage
	err := c.BindJSON(&standaloneMessage)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}
	// If any of the fields are empty, return an error
	if standaloneMessage.ConvoId == "" || standaloneMessage.Message.Prompt == "" || standaloneMessage.Message.Response == "" {
		c.JSON(400, gin.H{"error": "Missing fields"})
		return
	}
	// Append the message to the conversation messages array
	_, err = db.Exec(context.Background(), `UPDATE conversations SET messages = array_append(messages, $1) WHERE id = $2`, standaloneMessage.Message, standaloneMessage.ConvoId)
	if err != nil {
		err_code := 500
		err_string := err.Error()
		c.JSON(err_code, gin.H{"error": err_string})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}
