package lobby

type PayloadEvent string
type ResponseEvent string

const (
	UnknownPayloadEvent        PayloadEvent = "unknown"
	JoinPayloadEvent                        = "join_lobby"
	LeavePayloadEvent                       = "left_lobby"
	ChatPayloadEvent                        = "chat_message"
	ChooseColorPayloadEvent                 = "choose_color"
	SetReadyStatusPayloadEvent              = "set_ready_status"
)

const (
	UnknownResponseEvent        ResponseEvent = "unknown"
	JoinResponseEvent                         = "join_lobby"
	LeftResponseEvent                         = "left"
	NewMessageResponseEvent                   = "new_chat_message"
	ChooseColorResponseEvent                  = "choose_color"
	SetReadyStatusResponseEvent               = "set_ready_status"
	StartGameResponseEvent                    = "start_game"
)

