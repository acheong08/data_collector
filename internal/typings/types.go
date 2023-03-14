package typings

import (
	_ "github.com/jackc/pgx/v5"
)

type Message struct {
	Prompt   string `json:"prompt"`
	Response string `json:"response"`
}

type MessageInstance struct {
	Message Message `json:"message"`
	Id      string  `json:"id"`
	User    string  `json:"user"`
}
