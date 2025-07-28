package tgclient

import (
	"context"
	"log"

	"github.com/gotd/td/session"
)

// GenerateSampleSession generates sample session JSON and stores in provided session storage
func GenerateSampleSession(ctx context.Context, session *session.FileStorage) error {
	// sample session with minimally required session fields
	sessionJson := `{
  "Version": 1,
  "Data": {
	"DC": 0,
	"Addr": "",
	"AuthKey": "AUTH_KEY_HERE",
	"AuthKeyID": "AUTH_KEY_ID_HERE",
	"Salt": 12345
  }
}`
	err := session.StoreSession(ctx, []byte(sessionJson))
	if err != nil {
		log.Printf("failed to save sample sesion to file: %s", err)
		return err
	}
	log.Printf("sample session created, check %s", session.Path)
	return nil
}
