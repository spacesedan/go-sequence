package client

import (
	"bytes"
	"context"
	"fmt"

	"github.com/spacesedan/go-sequence/internal/lobby"
	"github.com/spacesedan/go-sequence/internal/views/components"
)

func (c *WsClient) handleJoin(r lobby.WsResponse) {
	var b bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if r.Sender != c.Username {
		components.PlayerStatus(r.Message).Render(ctx, &b)
		if err := c.sendResponse(b.String()); err != nil {
			fmt.Println("[ACTION] join_lobby", err.Error())
		c.errorChan <- err
		}
	}
	b.Reset()

	players, err := c.clientRepo.GetMPlayers(c.LobbyID, r.ConnectedUsers)
	if err != nil {
		c.errorChan <- err
	}
	components.PlayerDetails(players).Render(ctx, &b)
	if err := c.sendResponse(b.String()); err != nil {
		c.errorChan <- err
	}
	b.Reset()
}

// handleChatMessage handles incoming chat messages and send the correct
// component client based on the sender of the original message
func (c *WsClient) handleChatMessage(r lobby.WsResponse) {
	var b bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	alt := fmt.Sprintf("avatar image for %v", r.Sender)

	if r.Sender == c.Username {
		components.ChatMessageSender(
			r.Message,
			alt,
			generateUserAvatar(r.Sender, 32)).
			Render(ctx, &b)
	} else {
		components.ChatMessageReciever(
			r.Message,
			alt,
			generateUserAvatar(r.Sender, 32)).
			Render(ctx, &b)
	}
	if err := c.sendResponse(b.String()); err != nil {
        c.errorChan <- err
	}

	b.Reset()
}

func (c *WsClient) handleChooseColor(r lobby.WsResponse) {
	var b bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()


	sender, err := c.clientRepo.GetPlayer(c.LobbyID, r.Sender)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	err = components.PlayerUpdateDetails(sender).Render(ctx, &b)
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
	}

	if err := c.sendResponse(b.String()); err != nil {
		return
	}
	b.Reset()

}

func (c *WsClient) handlePlayerReady(r lobby.WsResponse) {
	var b bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if r.Sender == c.Username {
        ps, err := c.clientRepo.GetPlayer(c.LobbyID, c.Username)
        if err != nil {
            c.errorChan <- err
        }
		if ps.Color == "" {
			title := "Missing player color"
			content := "can't ready up without selecting a color"
			components.ToastWSComponent(title, content).Render(ctx, &b)
			c.sendResponse(b.String())
			b.Reset()
		}
	}

	sender, err := c.clientRepo.GetPlayer(c.LobbyID, r.Sender)
	if err != nil {
		return
	}

	components.PlayerUpdateDetails(sender).
		Render(context.Background(), &b)

	if err := c.sendResponse(b.String()); err != nil {
		return
	}
	b.Reset()
}
