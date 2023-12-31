package lobby

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spacesedan/go-sequence/internal"
)

type LobbyHandler interface {
	ChangeState()

	RegisterPlayer(WsPayload)
	DeregisterPlayer(WsPayload)

	DispatchAction(WsPayload)
	JoinAction(WsPayload)
	LeaveAction(WsPayload)
	ChatAction(WsPayload)
	ColorSelectionAction(WsPayload)
	ReadyAction(WsPayload)

	EmptyLobby() bool
}

type lobbyHandler struct {
	lobby  *Lobby
	logger *slog.Logger
	rdb    *redis.Client

	svc LobbyService
}

func NewLobbyHandler(r *redis.Client, l *Lobby, logger *slog.Logger) LobbyHandler {
	return &lobbyHandler{
		rdb:    r,
		lobby:  l,
		logger: logger,
		svc:    NewLobbyService(r, l, logger),
	}
}

func (h *lobbyHandler) RegisterPlayer(p WsPayload) {
	h.logger.Info("lobbyHandler.handleRegisterPlayer",
		fmt.Sprintf("player: %s joined lobby: %s", p.Username, h.lobby.ID), "OK")

	// check to see if the player is reconnecting or is a new player entirely
	var ps *internal.Player
	var r WsResponse

    time.Sleep(3 * time.Second)

	ps, err := h.svc.GetPlayer(p.Username)
	if ps == nil {
		ps = &internal.Player{
			Username: p.Username,
			LobbyId:  h.lobby.ID,
		}

		ps, err = h.svc.NewPlayer(p.Username)
		if err != nil {
			h.lobby.errorChan <- fmt.Errorf("handleRegisterPlayer error reason: %v", err)
		}

	}

	h.lobby.Players[p.Username] = ps
    h.svc.SetLobby(toLobbyState(h.lobby))

	// get the current state of the lobby
	cs := h.svc.GetCurrentState()

	// handle the register payload with the current state
	r.Sender = p.Username
	r.ConnectedUsers = h.svc.GetPlayerNames()

	switch cs {
	case internal.InLobby:
		r.Action = JoinLobbyResponseEvent
		r.Message = fmt.Sprintf("%s joined", p.Username)
	case internal.InGame:
		r.Action = JoinGameResponseEvent
		r.Message = fmt.Sprintf("%s reconnected", p.Username)
	}

	// send the response to the player
	h.publishResponse(r)

}

func (h *lobbyHandler) DeregisterPlayer(p WsPayload) {
	h.logger.Info("lobby.handleUnregisterSession",
		slog.Group("Unregistering player connection",
			slog.String("lobby_id", h.lobby.ID),
			slog.String("user", p.Username)))

	if _, ok := h.lobby.Players[p.Username]; ok {
		delete(h.lobby.Players, p.Username)
        // handle this in the lobby service
        // instead of calling to delete ill just remove the
        // the player from the the Player list and let the
        // unregistered player data to expire.
        // l.lobbyRepo.DeletePlayer(l.ID, payload.Username)
		h.svc.SetExpiration(p.Username, time.Duration(30*time.Second))
	}

}

func (h *lobbyHandler) ChangeState() {
	switch h.lobby.CurrentState.String() {
	case internal.InGame.String():

	}

}

func (h *lobbyHandler) DispatchAction(p WsPayload) {
	switch h.lobby.CurrentState {
	case internal.InLobby:
		switch p.Action {
		// case JoinLobbyPayloadEvent:
		// 	h.JoinAction(p)
		case LeavePayloadEvent:
			h.LeaveAction(p)
		case ChatPayloadEvent:
			h.ChatAction(p)
		case ChooseColorPayloadEvent:
			h.ColorSelectionAction(p)
		case SetReadyStatusPayloadEvent:
			h.ReadyAction(p)
		}
	case internal.InGame:
	}
}

func (h *lobbyHandler) JoinAction(p WsPayload) {
	var r WsResponse

	r.Action = JoinLobbyResponseEvent
	r.Message = fmt.Sprintf("%v joined", p.Username)
	r.SkipSender = true
	r.Sender = p.Username
	r.ConnectedUsers = h.svc.GetPlayerNames()

	h.publish(StateChannel, h.lobby.CurrentState)
	// if err := h.publish(StateChannel, h.lobby.CurrentState); err != nil {
	// 	h.lobby.errorChan <- err
	// }
}

func (h *lobbyHandler) LeaveAction(p WsPayload) {
	var r WsResponse

	r.Action = LeftResponseEvent
	r.Message = fmt.Sprintf("%v left", p.Username)
	r.SkipSender = true
	r.Sender = p.Username
	r.ConnectedUsers = h.svc.GetPlayerNames()

	if err := h.publishResponse(r); err != nil {
		h.lobby.errorChan <- err
	}
}

func (h *lobbyHandler) ChatAction(p WsPayload) {
	var r WsResponse

	r.Action = NewMessageResponseEvent
	r.Message = p.Message
	r.SkipSender = false
	r.Sender = p.Username
	r.ConnectedUsers = h.svc.GetPlayerNames()

	if err := h.publishResponse(r); err != nil {
		h.lobby.errorChan <- err
	}

}

func (h *lobbyHandler) ColorSelectionAction(p WsPayload) {
	var r WsResponse

	senderState, err := h.svc.GetPlayer(p.Username)
	if err != nil {
		h.lobby.errorChan <- fmt.Errorf("lobby.Subscribe err: ")
	}
	senderState.Color = p.Message
	h.svc.SetPlayer(senderState)

	h.lobby.Players[p.Username] = senderState
	h.svc.SetLobby(toLobbyState(h.lobby))

	r.Action = ChooseColorResponseEvent
	r.Sender = p.Username
	r.Message = p.Message
	r.ConnectedUsers = h.svc.GetPlayerNames()
	r.SkipSender = false
	if err := h.publishResponse(r); err != nil {
		h.lobby.errorChan <- err
	}

}

func (h *lobbyHandler) ReadyAction(p WsPayload) {
	var r WsResponse
	var playersReady []bool

	senderState, err := h.svc.GetPlayer(p.Username)
	if err != nil {
		h.lobby.errorChan <- err
	}
	senderState.Ready = true
	h.svc.SetPlayer(senderState)

	h.lobby.Players[p.Username] = senderState

	for _, p := range h.lobby.Players {
		if p.Ready {
			playersReady = append(playersReady, p.Ready)
		}
	}

	if len(playersReady) == h.lobby.Settings.NumOfPlayers {
		h.lobby.CurrentState = internal.InGame
		h.svc.SetLobby(toLobbyState(h.lobby))

        r.Action = JoinGameResponseEvent
        r.SkipSender = false
        r.ConnectedUsers = h.svc.GetPlayerNames()
        if err := h.publishResponse(r); err != nil {
            h.lobby.errorChan <- err
        }

	}

}

func (h *lobbyHandler) publish(c LobbyChannel, s internal.CurrentState) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := h.rdb.Publish(ctx, c.String(h.lobby.ID), WsPayload{
		Action:  "change_state",
		Message: s.String(),
	})
	if err.Err() != nil {
		h.lobby.errorChan <- err.Err()
	}

}

func (h *lobbyHandler) publishResponse(response WsResponse) error {
	h.logger.Info("lobby.publishResponse",
		slog.Group("sending response to players",
			slog.String("lobby_id", h.lobby.ID)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	responseChanKey := fmt.Sprintf("lobby.%v.responseChannel", h.lobby.ID)

	rb, err := response.MarshalBinary()
	if err != nil {
		h.logger.Error("wsClient.PublishPayloadToLobby",
			slog.Group("failed to marshal payload",
				slog.String("reason", err.Error())))
		return err
	}

	err = h.rdb.Publish(ctx, responseChanKey, rb).Err()
	if err != nil {
		h.logger.Error("lobby.publishResponse", slog.Group("error trying to publish", slog.String("lobby_id", h.lobby.ID)))
		return err
	}
	return nil
}

func (h *lobbyHandler) EmptyLobby() bool {
	if len(h.svc.GetPlayerNames()) == 0 {
		h.logger.Info("lobbyHandler.EmptyLobby",
			slog.Group("triggering closing lobby",
				slog.String("reason", "no players in lobby"),
				slog.String("lobby_id", h.lobby.ID),
			))
		return true
	}

	return false

}
