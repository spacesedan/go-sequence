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
			return
		}
	}
	b.Reset()

	components.PlayerDetails(r.ConnectedUsers).Render(ctx, &b)
	if err := c.sendResponse(b.String()); err != nil {
		return
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
		fmt.Println("[ACTION] new_chat_message", err.Error())
		return
	}

	b.Reset()
}

func (c *WsClient) handleChooseColor(r lobby.WsResponse) {
	var b bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()



	if c.Color == "" {
		c.setColor(r)
	} else {
		c.updateColor(r)
	}

	sender, err := c.clientRepo.GetPlayer(c.LobbyID, r.Sender)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	err = components.PlayerDetailsColored(
		sender.Username,
		r.Message,
		sender.Ready,
	).Render(ctx, &b)
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
		if c.Color == "" {
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

	components.PlayerDetailsColored(
		sender.Username,
		sender.Color,
		sender.Ready,
	).
		Render(context.Background(), &b)

	if err := c.sendResponse(b.String()); err != nil {
		return
	}
	b.Reset()
}
