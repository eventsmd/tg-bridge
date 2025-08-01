package tgsession

import (
	"encoding/json"
	"fmt"

	"github.com/gotd/td/session"
)

type ShortTelegramSession struct {
	Version int `json:"Version"`
	Data    struct {
		DC        int    `json:"DC"`
		Addr      string `json:"Addr"`
		AuthKey   []byte `json:"AuthKey"`
		AuthKeyID []byte `json:"AuthKeyID"`
		Salt      int64  `json:"Salt"`
	} `json:"Data"`
}

// TelegramSessionWrapper is gotd/td/session (jsonData) JSON representation
type TelegramSessionWrapper struct {
	Version int          `json:"Version"`
	Data    session.Data `json:"Data"`
}

// TrimSession converts a full Telegram session to a trimmed version (ShortTelegramSession)
// as session may further be encoded and encoded string is too long with original session.
// Trimmed session is compatible with Telegram API client.
func TrimSession(fullSessionJSON []byte) ([]byte, error) {
	var fullSession TelegramSessionWrapper
	if err := json.Unmarshal(fullSessionJSON, &fullSession); err != nil {
		return nil, fmt.Errorf("failed to unmarshal full session: %w", err)
	}

	shortSession := ShortTelegramSession{
		Version: fullSession.Version,
	}
	shortSession.Data.DC = fullSession.Data.DC
	shortSession.Data.Addr = fullSession.Data.Addr
	shortSession.Data.AuthKey = fullSession.Data.AuthKey
	shortSession.Data.AuthKeyID = fullSession.Data.AuthKeyID
	shortSession.Data.Salt = fullSession.Data.Salt

	shortSessionJSON, err := json.MarshalIndent(shortSession, "", "  ")

	return shortSessionJSON, err
}
