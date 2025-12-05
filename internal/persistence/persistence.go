package persistence

import (
	"context"
	"tg-bridge/internal/domain"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DatabaseConnection struct {
	pool *pgxpool.Pool
}

func NewDatabase(connString string) (*DatabaseConnection, error) {
	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	db := &DatabaseConnection{pool: pool}

	_, err = db.pool.Exec(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS last_message_offsets (
			chat_id NUMERIC PRIMARY KEY,
			last_message_id NUMERIC NOT NULL
		)`,
	)
	if err != nil {
		pool.Close()
		return nil, err
	}

	return db, nil
}

func (c *DatabaseConnection) SaveLastMessageID(messages []domain.Message) error {
	if len(messages) == 0 {
		return nil
	}
	ctx := context.Background()

	maxPerChat := make(map[domain.ChatID]domain.MessageID, len(messages))
	for _, m := range messages {
		if m.ID == 0 || m.ChatID == 0 {
			continue
		}
		if cur, ok := maxPerChat[m.ChatID]; !ok || m.ID > cur {
			maxPerChat[m.ChatID] = m.ID
		}
	}
	if len(maxPerChat) == 0 {
		return nil
	}

	const q = `
		INSERT INTO last_message_offsets (chat_id, last_message_id)
		VALUES ($1, $2)
		ON CONFLICT (chat_id)
		DO UPDATE SET last_message_id = GREATEST(EXCLUDED.last_message_id, last_message_offsets.last_message_id)
	`

	batch := &pgx.Batch{}
	for chatID, maxID := range maxPerChat {
		batch.Queue(q, int64(chatID), int64(maxID))
	}

	br := c.pool.SendBatch(ctx, batch)
	defer func(br pgx.BatchResults) {
		_ = br.Close()
	}(br)

	for range maxPerChat {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (c *DatabaseConnection) GetLastMessageID(channel domain.ChatID) (domain.MessageID, error) {
	ctx := context.Background()
	var last int64
	err := c.pool.QueryRow(ctx, `
		SELECT last_message_id
		FROM last_message_offsets
		WHERE chat_id = $1
	`, int64(channel)).Scan(&last)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return domain.MessageID(last), nil
}

func (c *DatabaseConnection) Close() {
	c.pool.Close()
}
