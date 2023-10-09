package lobby

import "github.com/gorilla/websocket"

type WsConnection struct {
	*websocket.Conn
}

type WsJsonResponse struct {
	Headers        map[string]interface{} `json:"HEADERS"`
	Action         string                 `json:"action"`
	Message        string                 `json:"message"`
	LobbyID        string
	MessageType    string       `json:"message_type"`
	SkipSender     bool         `json:"-"`
	CurrentConn    WsConnection `json:"-"`
	ConnectedUsers []string     `json:"-"`
}

type WsPayload struct {
	Headers  map[string]string `json:"HEADERS"`
	Action   string            `json:"action"`
	Settings Settings          `json:"settings"`
	ID       string            `json:"id"`
	LobbyID  string            `json:"lobby_id"`
	User     string            `json:"user"`
	Message  string            `json:"message"`
	Conn     WsConnection
}
