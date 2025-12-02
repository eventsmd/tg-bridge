package persistence

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"tg-bridge/internal/domain"

	"github.com/jackc/pgx/v4"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func startPostgres(t *testing.T) (*tcpostgres.PostgresContainer, string) {
	t.Helper()

	ctx := context.Background()
	container, err := tcpostgres.Run(ctx,
		"postgres:17",
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	if container == nil {
		t.Fatalf("failed to start postgres container")
	}
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}
	return container, connStr
}

func Test_SaveAndGetLastMessageID(t *testing.T) {
	ctx := context.Background()
	container, connStr := startPostgres(t)
	defer func() {
		_ = container.Terminate(context.Background())
	}()

	db, err := NewDatabase(connStr)
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	type step struct {
		msgs   []domain.Message
		checks map[domain.ChatID]domain.MessageID
	}

	steps := []step{
		{
			// initial insert for chat 1
			msgs: []domain.Message{
				{ID: 10, ChatID: 1},
			},
			checks: map[domain.ChatID]domain.MessageID{
				1: 11,
			},
		},
		{
			// lower id should not decrease stored value
			msgs: []domain.Message{
				{ID: 5, ChatID: 1},
			},
			checks: map[domain.ChatID]domain.MessageID{
				1: 11,
			},
		},
		{
			// higher id should update
			msgs: []domain.Message{
				{ID: 12, ChatID: 1},
			},
			checks: map[domain.ChatID]domain.MessageID{
				1: 13,
			},
		},
		{
			// another chat independent
			msgs: []domain.Message{
				{ID: 7, ChatID: 2},
			},
			checks: map[domain.ChatID]domain.MessageID{
				1: 13,
				2: 8,
			},
		},
		{
			// multiple chats in one batch
			msgs: []domain.Message{
				{ID: 15, ChatID: 1},
				{ID: 9, ChatID: 2},
				{ID: 3, ChatID: 3},
			},
			checks: map[domain.ChatID]domain.MessageID{
				1: 16,
				2: 10,
				3: 4,
			},
		},
		{
			// multiple messages in one chat
			msgs: []domain.Message{
				{ID: 24, ChatID: 29},
				{ID: 120, ChatID: 29},
				{ID: 15, ChatID: 29},
			},
			checks: map[domain.ChatID]domain.MessageID{
				29: 121,
			},
		},
	}

	for i, st := range steps {
		t.Run(fmt.Sprintf("step_%d", i+1), func(t *testing.T) {
			if err := db.SaveLastMessageID(st.msgs); err != nil {
				t.Fatalf("SaveLastMessageID failed: %v", err)
			}
			for chat, want := range st.checks {
				got, err := db.GetLastMessageID(chat)
				if err != nil {
					t.Fatalf("GetLastMessageID(%d) error: %v", chat, err)
				}
				if got != want {
					t.Fatalf("GetLastMessageID(%d) = %d, want %d", chat, got, want)
				}
			}
		})
	}

	t.Run("unknown_chat_returns_zero", func(t *testing.T) {
		got, err := db.GetLastMessageID(9999)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Fatalf("expected 0 for unknown chat, got %d", got)
		}
	})

	// idempotency: reapply the same max value
	t.Run("idempotent_update", func(t *testing.T) {
		msg := domain.Message{ID: 15, ChatID: 1}
		if err := db.SaveLastMessageID([]domain.Message{msg}); err != nil {
			t.Fatalf("SaveLastMessageID failed: %v", err)
		}
		got, err := db.GetLastMessageID(1)
		if err != nil {
			t.Fatalf("GetLastMessageID error: %v", err)
		}
		if got != 16 {
			t.Fatalf("expected 16, got %d", got)
		}
	})

	_ = ctx // reserved for potential future context usage
}
