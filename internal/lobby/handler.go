package lobby

import (
	"fmt"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/spacesedan/go-sequence/internal"
)

type LobbyHandler interface {
	RegisterPlayer(WsPayload)
    DeregisterPlayer(WsPayload)
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

	ps := &internal.Player{
		Username: p.Username,
		LobbyId:  h.lobby.ID,
	}

	ps, err := h.svc.NewPlayer(ps)
	if err != nil {
		h.lobby.errorChan <- fmt.Errorf("handleRegisterPlayer error reason: %v", err)
	}

	h.lobby.Players[p.Username] = ps

}

func (h *lobbyHandler) DeregisterPlayer(p WsPayload) {
	h.logger.Info("lobby.handleUnregisterSession",
		slog.Group("Unregistering player connection",
			slog.String("lobby_id", h.lobby.ID),
			slog.String("user", p.Username)))

	if _, ok := h.lobby.Players[p.Username]; ok {
		delete(h.lobby.Players, p.Username)
        h.svc.ShortenPlayerExpiration(p.Username)
        // handle this in the lobby service
		// instead of calling to delete ill just remove the
		// the player from the the Player list and let the
		// unregistered player data to expire.
		// l.lobbyRepo.DeletePlayer(l.ID, payload.Username)
	}

}
