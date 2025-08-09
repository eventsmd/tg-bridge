package temporalpub

import (
	"context"
	"fmt"
	"log/slog"
	"tg-bridge/internal/config"
	"tg-bridge/internal/domain"
	"time"

	"go.temporal.io/sdk/client"
)

type Publisher struct {
	tc     client.Client
	cfg    config.Config
	logger *slog.Logger
}

func NewPublisher(cfg config.Config, logger *slog.Logger, existing client.Client) (*Publisher, error) {
	if logger == nil {
		logger = slog.Default()
	}
	var (
		tc  client.Client
		err error
	)
	if existing != nil {
		tc = existing
	} else {
		tc, err = client.Dial(client.Options{

			HostPort:  cfg.TemporalHostPort,
			Namespace: cfg.TemporalNamespace,
		})
		if err != nil {
			return nil, fmt.Errorf("temporal dial: %w", err)
		}
	}
	return &Publisher{
		tc:     tc,
		cfg:    cfg,
		logger: logger,
	}, nil
}

// Close — закрыть клиент при завершении работы приложения
func (p *Publisher) Close() error {
	if p.tc != nil {
		p.tc.Close()
	}
	return nil
}

func (p *Publisher) StartTelegramWorkflow(ctx context.Context, msg domain.Message) (workflowID, runID string, err error) {
	if p.tc == nil {
		return "", "", fmt.Errorf("temporal client is not initialized")
	}
	wfID := p.workflowIDFor(msg)

	opts := client.StartWorkflowOptions{
		ID:                       wfID,
		TaskQueue:                p.cfg.TemporalTaskQueue,
		WorkflowExecutionTimeout: 24 * time.Hour,
	}

	run, err := p.tc.ExecuteWorkflow(ctx, opts, p.cfg.TemporalWorkflowType, msg)
	if err != nil {
		return "", "", fmt.Errorf("execute workflow: %w", err)
	}
	return run.GetID(), run.GetRunID(), nil
}

func (p *Publisher) workflowIDFor(msg domain.Message) string {
	return fmt.Sprintf("tg:%d:%d", msg.ChatID, msg.ID)
}
