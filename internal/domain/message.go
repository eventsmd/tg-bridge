package domain

import (
	"encoding/json"
	"errors"
	"time"
)

type (
	ChatID    int64
	UserID    int64
	MessageID int64
)

type User struct {
	ID   UserID `json:"id"`
	Name string `json:"name,omitempty"`
}

type Supplier struct {
	Type string `json:"type"`
}

type MessageRef struct {
	ID     MessageID `json:"id"`
	ChatID ChatID    `json:"chat_id"`
}

type Message struct {
	ID      MessageID      `json:"id"`
	ChatID  ChatID         `json:"chat_id"`
	From    User           `json:"from"`
	Text    string         `json:"text"`
	Date    time.Time      `json:"date"`
	ReplyTo *MessageRef    `json:"reply_to,omitempty"`
	Context map[string]any `json:"context,omitempty"`
}

func NewMessage(
	id MessageID,
	chatID ChatID,
	from User,
	text string,
	date time.Time,
	replyTo *MessageRef,
	context map[string]any,
) (Message, error) {
	if id == 0 {
		return Message{}, errors.New("message id is required")
	}
	if chatID == 0 {
		return Message{}, errors.New("chat id is required")
	}
	if from.ID == 0 {
		return Message{}, errors.New("from user id is required")
	}
	if date.IsZero() {
		date = time.Now().UTC()
	}
	return Message{
		ID:      id,
		ChatID:  chatID,
		From:    from,
		Text:    text,
		Date:    date.UTC(),
		ReplyTo: replyTo,
		Context: context,
	}, nil
}

func (m Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
