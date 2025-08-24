package tgclient

import (
	"context"
	"fmt"
	"tg-bridge/internal/domain"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type Channel struct {
	client   *telegram.Client
	channel  *tg.Channel
	supplier domain.Supplier
}

func NewChannel(ctx context.Context, client *telegram.Client, name string, supplier domain.Supplier) (*Channel, error) {

	api := client.API()

	resolved, err := api.ContactsResolveUsername(ctx,
		&tg.ContactsResolveUsernameRequest{
			Username: name,
		})
	if err != nil {
		err := fmt.Errorf("failed to resolve channel username: %v", err)
		return nil, err
	}

	if len(resolved.Chats) == 1 {

		channel, converted := resolved.Chats[0].(*tg.Channel)

		if !converted {
			return nil, fmt.Errorf("is not a channel: %v", name)
		}

		return &Channel{
			client:   client,
			channel:  channel,
			supplier: supplier,
		}, nil
	}
	return nil, fmt.Errorf("channels found: %v", len(resolved.Chats))
}

func (c *Channel) Messages(ctx context.Context, limit int, offset int) ([]domain.Message, error) {
	hist, err := c.client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:     c.channel.AsInputPeer(),
		Limit:    limit,
		OffsetID: offset,
	})
	if err != nil {
		return nil, err
	}

	channelMsgs, ok := hist.(*tg.MessagesChannelMessages)
	if !ok {
		return nil, fmt.Errorf("unexpected history type: %T", hist)
	}

	result := make([]domain.Message, 0, len(channelMsgs.Messages))

	for _, obj := range channelMsgs.Messages {
		msg, ok := obj.(*tg.Message)
		if !ok {
			continue
		}

		var fromUserID int64
		if pu, ok := msg.FromID.(*tg.PeerUser); ok {
			fromUserID = pu.UserID
		}

		var reply *domain.MessageRef
		if msg.ReplyTo != nil {
			if msgReply, ok := msg.ReplyTo.(*tg.MessageReplyHeader); ok {
				reply = &domain.MessageRef{
					ID:     domain.MessageID(msgReply.ReplyToMsgID),
					ChatID: domain.ChatID(c.channel.ID),
				}
			}
		}

		ctxMap := map[string]any{
			"supplier": c.supplier.Type,
		}

		newMess, err := domain.NewMessage(
			domain.MessageID(msg.ID),
			domain.ChatID(c.channel.ID),
			domain.User{
				ID:   domain.UserID(fromUserID),
				Name: msg.PostAuthor,
			},
			msg.Message,
			time.Unix(int64(msg.Date), 0).UTC(),
			reply,
			ctxMap,
		)

		if err != nil {
			return nil, err
		}

		result = append(
			result,
			newMess,
		)
	}

	return result, nil
}

func (c *Channel) Id() int64 {
	return c.channel.ID
}
