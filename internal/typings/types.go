package typings

type Message struct {
	Prompt   string `json:"prompt"`
	Response string `json:"response"`
}

type Conversation struct {
	Messages []Message `json:"messages"`
	Id       string    `json:"id"`
	User     string    `json:"user"`
}
