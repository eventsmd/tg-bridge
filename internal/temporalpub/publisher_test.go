package temporalpub

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"tg-bridge/internal/config"
	"tg-bridge/internal/domain"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const (
	temporalImage = "temporalio/auto-setup:latest"
	postgresImage = "postgres:15-alpine"
)

func testWorkflow(_ workflow.Context, msg domain.Message) (string, error) {
	return fmt.Sprintf("ok:%d", msg.ID), nil
}

func TestPublisher_StartTelegramWorkflow_E2E_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()

	// 0) Isolated network
	netw := createNetwork(ctx, t)
	t.Cleanup(func() { _ = netw.Remove(context.Background()) })

	// 1) Postgres
	pgC := startPostgres(ctx, t, netw)
	t.Cleanup(func() { _ = pgC.Terminate(context.Background()) })

	// 2) Temporal + wait for namespace ready
	temporalC, hostPort := startTemporal(ctx, t, netw)
	t.Cleanup(func() { _ = temporalC.Terminate(context.Background()) })
	if err := waitNamespaceReady(ctx, hostPort, "default"); err != nil {
		dumpContainerLogs(t, ctx, "temporal", temporalC)
		t.Fatalf("namespace not ready: %v", err)
	}

	// 3) SDK client
	cli := mustDialTemporal(t, hostPort, "default")
	defer cli.Close()

	// 4) Worker
	cfg := config.Config{
		TemporalHostPort:     hostPort,
		TemporalNamespace:    "default",
		TemporalTaskQueue:    "telegram-workflows-e2e",
		TemporalWorkflowType: "TelegramMessageWorkflowTest",
	}
	w := startWorker(t, cli, cfg.TemporalTaskQueue, cfg.TemporalWorkflowType)
	defer w.Stop()

	// 5) Publisher
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	pub, err := NewPublisher(cfg, logger, nil)
	if err != nil {
		t.Fatalf("new publisher: %v", err)
	}
	defer func(pub *Publisher) {
		_ = pub.Close()
	}(pub)

	// 6) Start a workflow and get a result
	msg := domain.Message{
		ID:     1001,
		ChatID: 2002,
		From: domain.User{
			ID:   3003,
			Name: "tester",
		},
		Text:    "hello from test",
		Date:    time.Now().UTC(),
		Context: map[string]any{"source": "e2e"},
	}

	wfID, runID, err := pub.StartTelegramWorkflow(ctx, msg)
	if err != nil {
		dumpContainerLogs(t, ctx, "temporal", temporalC)
		t.Fatalf("start workflow: %v", err)
	}
	if wfID == "" || runID == "" {
		t.Fatalf("empty workflow/run id")
	}

	run := cli.GetWorkflow(ctx, wfID, runID)
	var result string
	if err := run.Get(ctx, &result); err != nil {
		dumpContainerLogs(t, ctx, "temporal", temporalC)
		t.Fatalf("get workflow result: %v", err)
	}
	want := fmt.Sprintf("ok:%d", msg.ID)
	if result != want {
		t.Fatalf("unexpected result: got %q, want %q", result, want)
	}
}

// Helpers

func createNetwork(ctx context.Context, t *testing.T) *testcontainers.DockerNetwork {
	t.Helper()
	netw, err := network.New(ctx)
	if err != nil {
		t.Fatalf("create network: %v", err)
	}
	return netw
}

func startPostgres(ctx context.Context, t *testing.T, netw *testcontainers.DockerNetwork) testcontainers.Container {
	t.Helper()
	const (
		pgUser = "temporal"
		pgPass = "temporal"
		pgDB   = "temporal"
		pgHost = "postgres"
		pgPort = "5432"
	)

	req := testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{pgPort + "/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     pgUser,
			"POSTGRES_PASSWORD": pgPass,
			"POSTGRES_DB":       pgDB,
		},
		Networks:       []string{netw.Name},
		NetworkAliases: map[string][]string{netw.Name: {pgHost}},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(pgPort+"/tcp"),
			wait.ForLog("database system is ready to accept connections"),
		),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	return c
}

func startTemporal(ctx context.Context, t *testing.T, netw *testcontainers.DockerNetwork) (testcontainers.Container, string) {
	t.Helper()
	const (
		pgUser = "temporal"
		pgPass = "temporal"
		pgDB   = "temporal"
		pgHost = "postgres"
		pgPort = "5432"
	)

	req := testcontainers.ContainerRequest{
		Image:        temporalImage,
		ExposedPorts: []string{"7233/tcp"},
		Env: map[string]string{
			"DB":                "postgres12",
			"POSTGRES_SEEDS":    pgHost,
			"DB_PORT":           pgPort,
			"POSTGRES_USER":     pgUser,
			"POSTGRES_PASSWORD": pgPass,
			"POSTGRES_PWD":      pgPass,
			"POSTGRES_DB":       pgDB,
			"DEFAULT_NAMESPACE": "default",
			"ENABLE_ES":         "false",
		},
		Networks: []string{netw.Name},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("7233/tcp"),
		),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          false,
	})
	if err != nil {
		t.Fatalf("create temporal container: %v", err)
	}
	if err := c.Start(ctx); err != nil {
		dumpContainerLogs(t, ctx, "temporal", c)
		t.Fatalf("start temporal container: %v", err)
	}
	host, err := c.Host(ctx)
	if err != nil {
		dumpContainerLogs(t, ctx, "temporal", c)
		t.Fatalf("container host: %v", err)
	}
	mapped, err := c.MappedPort(ctx, "7233/tcp")
	if err != nil {
		dumpContainerLogs(t, ctx, "temporal", c)
		t.Fatalf("mapped port: %v", err)
	}
	return c, fmt.Sprintf("%s:%s", host, mapped.Port())
}

func waitNamespaceReady(ctx context.Context, address, namespace string) error {
	waitCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	// 1) SDK client with retries
	var (
		cli client.Client
		err error
	)
	backoff := 500 * time.Millisecond
	for {
		cli, err = client.Dial(client.Options{
			HostPort:  address,
			Namespace: namespace,
		})
		if err == nil {
			break
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("temporal dial: %w", waitCtx.Err())
		case <-time.After(backoff):
			if backoff < 5*time.Second {
				backoff *= 2
				if backoff > 5*time.Second {
					backoff = 5 * time.Second
				}
			}
		}
	}
	defer cli.Close()

	// 2) Ждем, пока namespace будет доступен
	svc := cli.WorkflowService()
	backoff = 500 * time.Millisecond
	for {
		_, err = svc.DescribeNamespace(waitCtx, &workflowservice.DescribeNamespaceRequest{
			Namespace: namespace,
		})
		if err == nil {
			return nil
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("namespace %q not ready: %w", namespace, waitCtx.Err())
		case <-time.After(backoff):
			if backoff < 5*time.Second {
				backoff *= 2
				if backoff > 5*time.Second {
					backoff = 5 * time.Second
				}
			}
		}
	}
}

func mustDialTemporal(t *testing.T, address, namespace string) client.Client {
	t.Helper()
	cli, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		t.Fatalf("temporal dial: %v", err)
	}
	return cli
}

func startWorker(t *testing.T, cli client.Client, taskQueue, workflowType string) worker.Worker {
	t.Helper()
	w := worker.New(cli, taskQueue, worker.Options{})
	w.RegisterWorkflowWithOptions(testWorkflow, workflow.RegisterOptions{Name: workflowType})
	if err := w.Start(); err != nil {
		t.Fatalf("worker start: %v", err)
	}
	return w
}

func dumpContainerLogs(t *testing.T, ctx context.Context, name string, c testcontainers.Container) {
	t.Helper()
	if c == nil {
		return
	}
	rc, err := c.Logs(ctx)
	if err != nil {
		t.Logf("[%s] logs: unable to get logs: %v", name, err)
		return
	}
	defer func(rc io.ReadCloser) {
		_ = rc.Close()
	}(rc)
	b, _ := io.ReadAll(rc)
	if len(b) == 0 {
		t.Logf("[%s] logs: (empty)", name)
	} else {
		t.Logf("[%s] container logs:\n%s", name, string(b))
	}
}
