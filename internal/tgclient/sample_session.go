package tgclient

import (
	"context"
	"log"
	"tg-bridge/internal/tgsession"
)

// GenerateSampleSession generates sample session JSON and stores in provided session storage
func GenerateSampleSession(ctx context.Context, session *tgsession.MemorySessionStorage) error {
	log.Printf("Generating sample session")
	// sample session with minimally required session fields
	sessionJson := `{
  "Version": 1,
  "Data": {
	"DC": 0,
	"Addr": "",
	"AuthKey": "QVVUSF9LRVlfSEVSRQo=",
	"AuthKeyID": "QVVUSF9LRVlfSURfSEVSRQo=",
	"Salt": 12345
  }
}`
	err := session.StoreSession(ctx, []byte(sessionJson))
	if err != nil {
		log.Printf("failed to save sample sesion: %s", err)
		return err
	}
	return PrintEncodedSession(ctx, session)
}
