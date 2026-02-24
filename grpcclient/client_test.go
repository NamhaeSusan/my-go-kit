package grpcclient

import (
	"testing"
	"time"

	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func TestCheckClientConfigSetsDefaults(t *testing.T) {
	cfg := Config{}

	checkClientConfig(&cfg)

	if cfg.TransportCredentials == nil {
		t.Fatal("transport credentials should be set")
	}
	if cfg.ReadBufferSize != defaultReadBufferSize {
		t.Fatalf("unexpected read buffer size: %d", cfg.ReadBufferSize)
	}
	if cfg.WriteBufferSize != defaultWriteBufferSize {
		t.Fatalf("unexpected write buffer size: %d", cfg.WriteBufferSize)
	}
	if cfg.MaxRecvMsgSize != defaultMaxRecvMsgSize {
		t.Fatalf("unexpected max recv message size: %d", cfg.MaxRecvMsgSize)
	}
	if cfg.MaxSendMsgSize != defaultMaxSendMsgSize {
		t.Fatalf("unexpected max send message size: %d", cfg.MaxSendMsgSize)
	}
	if cfg.InitialWindowSize != defaultInitialWindowSize {
		t.Fatalf("unexpected initial window size: %d", cfg.InitialWindowSize)
	}
	if cfg.InitialConnWindowSize != defaultInitialConnWindowSize {
		t.Fatalf("unexpected initial connection window size: %d", cfg.InitialConnWindowSize)
	}
	if cfg.KeepAliveClientParameters == nil {
		t.Fatal("keepalive params should be set")
	}
	if cfg.IdleTimeout != defaultIdleTimeout {
		t.Fatalf("unexpected idle timeout: %s", cfg.IdleTimeout)
	}
}

func TestCheckClientConfigPreservesProvidedValues(t *testing.T) {
	creds := insecure.NewCredentials()
	ka := &keepalive.ClientParameters{
		Time:                2 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: false,
	}
	cfg := Config{
		TransportCredentials:      &creds,
		ReadBufferSize:            1,
		WriteBufferSize:           2,
		MaxRecvMsgSize:            3,
		MaxSendMsgSize:            4,
		InitialWindowSize:         5,
		InitialConnWindowSize:     6,
		KeepAliveClientParameters: ka,
		IdleTimeout:               7 * time.Second,
	}

	checkClientConfig(&cfg)

	if cfg.TransportCredentials != &creds {
		t.Fatal("transport credentials should not be replaced")
	}
	if cfg.ReadBufferSize != 1 || cfg.WriteBufferSize != 2 {
		t.Fatalf("buffer sizes should be preserved: read=%d write=%d", cfg.ReadBufferSize, cfg.WriteBufferSize)
	}
	if cfg.MaxRecvMsgSize != 3 || cfg.MaxSendMsgSize != 4 {
		t.Fatalf("message sizes should be preserved: recv=%d send=%d", cfg.MaxRecvMsgSize, cfg.MaxSendMsgSize)
	}
	if cfg.InitialWindowSize != 5 || cfg.InitialConnWindowSize != 6 {
		t.Fatalf("window sizes should be preserved: win=%d connWin=%d", cfg.InitialWindowSize, cfg.InitialConnWindowSize)
	}
	if cfg.KeepAliveClientParameters != ka {
		t.Fatal("keepalive params should not be replaced")
	}
	if cfg.IdleTimeout != 7*time.Second {
		t.Fatalf("timeouts should be preserved: idle=%s", cfg.IdleTimeout)
	}
}

func TestNewClientAndRoundRobinGetConn(t *testing.T) {
	client, err := NewClient("passthrough:///unit-test", Config{})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer func() { _ = client.Close() }()

	if len(client.conns) != defaultMaxConnections {
		t.Fatalf("unexpected number of pooled connections: %d", len(client.conns))
	}

	// index starts from 0, first AddUint64 makes it 1.
	expectedOrder := []int{1, 2, 3, 0, 1}
	for i, expected := range expectedOrder {
		got := client.GetConn()
		if got != client.conns[expected] {
			t.Fatalf("unexpected connection at step %d: expected index %d", i, expected)
		}
	}
}

func TestReconnectReplacesMatchedConnection(t *testing.T) {
	client, err := NewClient("passthrough:///unit-test", Config{})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer func() { _ = client.Close() }()

	oldConn := client.conns[0]
	if err := client.Reconnect(oldConn); err != nil {
		t.Fatalf("Reconnect returned error: %v", err)
	}

	if client.conns[0] == oldConn {
		t.Fatal("connection should be replaced after reconnect")
	}
}

func TestReconnectIgnoresUnknownConnection(t *testing.T) {
	client, err := NewClient("passthrough:///unit-test", Config{})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer func() { _ = client.Close() }()

	before := client.conns[0]
	if err := client.Reconnect(nil); err != nil {
		t.Fatalf("Reconnect should ignore unknown connection, got error: %v", err)
	}
	if client.conns[0] != before {
		t.Fatal("connection should not change when old connection is not in pool")
	}
}
