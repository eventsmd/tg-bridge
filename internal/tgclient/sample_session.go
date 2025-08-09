package tgclient

import (
	"context"
	"github.com/fatih/color"
	"tg-bridge/internal/tgsession"
)

// GenerateSampleSession generates sample session JSON and stores in provided session storage
func GenerateSampleSession(ctx context.Context, session *tgsession.MemorySessionStorage) error {
	_, _ = color.New(color.FgBlack, color.BgHiYellow).Print(" Warning! ")
	color.Yellow(" Generating sample session...")
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
		color.Red("failed to save sample session: %s", err)
		return err
	}
	sessionBase64, err := GetEncodedSession(ctx, session)
	if err == nil {
		color.Cyan("Base64 encoded session:")
		color.Green(sessionBase64)
	}
	return err
}
