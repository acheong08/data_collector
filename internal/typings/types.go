package typings

type Message struct {
	Prompt   string `json:"prompt"`
	Response string `json:"response"`
	ConvoId  string `json:"convo_id"`
}

type MessageInstance struct {
	Message Message `json:"message"`
	Id      string  `json:"id"`
	User    string  `json:"user"`
}
