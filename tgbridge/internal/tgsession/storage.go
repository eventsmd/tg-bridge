package tgsession

import (
	"context"
	"encoding/base64"
	"log"
	"tg-bridge/tgbridge/internal/config"

	"github.com/gotd/td/session"
)

type MemorySessionStorage struct {
	sessionData []byte
}

func NewMemorySessionStorage(initialData []byte) *MemorySessionStorage {
	return &MemorySessionStorage{sessionData: initialData}
}

func (s *MemorySessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	if len(s.sessionData) == 0 {
		return nil, session.ErrNotFound
	}
	return s.sessionData, nil
}

func (s *MemorySessionStorage) StoreSession(ctx context.Context, data []byte) error {
	s.sessionData = data
	return nil
}

// CreateSessionStorage Creates in memory session storage with previously generated telegram session
func CreateSessionStorage(cfg config.Config) (session.Storage, error) {
	var sessionStorage session.Storage
	sessionData, err := base64.StdEncoding.DecodeString(cfg.TelegramSession)
	if err != nil {
		log.Fatal("Invalid base64 session data:", err)
	}
	sessionStorage = NewMemorySessionStorage(sessionData)
	return sessionStorage, err
}
