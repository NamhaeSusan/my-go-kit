package grpcclient

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NamhaeSusan/my-go-kit/grpcclient/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultMaxConnections        = 4
	defaultReadBufferSize        = 1024 * 32
	defaultWriteBufferSize       = 1024 * 32
	defaultMaxRecvMsgSize        = 1024 * 1024 * 4
	defaultMaxSendMsgSize        = 1024 * 1024 * 4
	defaultInitialWindowSize     = 1024 * 1024 * 1
	defaultInitialConnWindowSize = 1024 * 1024 * 2
)

var (
	defaultKeepAliveClientParameters = &keepalive.ClientParameters{
		Time:                1 * time.Second,
		Timeout:             30 * time.Second,
		PermitWithoutStream: true,
	}
	defaultIdleTimeout          = 600 * time.Second
	defaultTransportCredentials = insecure.NewCredentials()
)

type Config struct {
	MaxConnections            int
	TransportCredentials      *credentials.TransportCredentials
	ReadBufferSize            int
	WriteBufferSize           int
	MaxRecvMsgSize            int
	MaxSendMsgSize            int
	InitialWindowSize         int32
	InitialConnWindowSize     int32
	KeepAliveClientParameters *keepalive.ClientParameters
	IdleTimeout               time.Duration
	UnaryClientInterceptors   []grpc.UnaryClientInterceptor
	StreamClientInterceptors  []grpc.StreamClientInterceptor
}

type Client struct {
	conns       []*grpc.ClientConn
	index       uint64
	mu          sync.RWMutex
	addr        string
	dialOptions []grpc.DialOption
}

func checkClientConfig(cfg *Config) {
	if cfg.MaxConnections <= 0 {
		cfg.MaxConnections = defaultMaxConnections
	}
	if cfg.TransportCredentials == nil {
		cfg.TransportCredentials = &defaultTransportCredentials
	}
	if cfg.ReadBufferSize <= 0 {
		cfg.ReadBufferSize = defaultReadBufferSize
	}
	if cfg.WriteBufferSize <= 0 {
		cfg.WriteBufferSize = defaultWriteBufferSize
	}
	if cfg.MaxRecvMsgSize <= 0 {
		cfg.MaxRecvMsgSize = defaultMaxRecvMsgSize
	}
	if cfg.MaxSendMsgSize <= 0 {
		cfg.MaxSendMsgSize = defaultMaxSendMsgSize
	}
	if cfg.InitialWindowSize <= 0 {
		cfg.InitialWindowSize = defaultInitialWindowSize
	}
	if cfg.InitialConnWindowSize <= 0 {
		cfg.InitialConnWindowSize = defaultInitialConnWindowSize
	}
	if cfg.KeepAliveClientParameters == nil {
		cfg.KeepAliveClientParameters = defaultKeepAliveClientParameters
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = defaultIdleTimeout
	}
}

func NewClient(addr string, cfg Config) (*Client, error) {
	checkClientConfig(&cfg)

	unaryInterceptors := append([]grpc.UnaryClientInterceptor{
		interceptor.UnaryClientTraceInterceptor(),
		interceptor.UnaryClientLoggingInterceptor(),
	}, cfg.UnaryClientInterceptors...)

	streamInterceptors := append([]grpc.StreamClientInterceptor{
		interceptor.StreamClientTraceInterceptor(),
		interceptor.StreamClientLoggingInterceptor(),
	}, cfg.StreamClientInterceptors...)

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(*cfg.TransportCredentials),
		grpc.WithReadBufferSize(cfg.ReadBufferSize),
		grpc.WithWriteBufferSize(cfg.WriteBufferSize),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(cfg.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(cfg.MaxSendMsgSize),
		),
		grpc.WithInitialWindowSize(cfg.InitialWindowSize),
		grpc.WithInitialConnWindowSize(cfg.InitialConnWindowSize),
		grpc.WithKeepaliveParams(*cfg.KeepAliveClientParameters),
		grpc.WithIdleTimeout(cfg.IdleTimeout),
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
		grpc.WithChainStreamInterceptor(streamInterceptors...),
		// Note: 밑으로는 default 세팅
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  100 * time.Millisecond,
				MaxDelay:   3 * time.Second,
				Multiplier: 1.6,
			},
			MinConnectTimeout: 3 * time.Second,
		}),

		grpc.WithDefaultServiceConfig(
			`{
				"loadBalancingConfig": [{"round_robin":{}}]
			}`,
		),
	}

	client := &Client{
		addr:        addr,
		dialOptions: dialOptions,
	}

	grpcConnections := make([]*grpc.ClientConn, cfg.MaxConnections)
	for i := range grpcConnections {
		conn, err := grpc.NewClient(addr, dialOptions...)
		if err != nil {
			return nil, err
		}
		grpcConnections[i] = conn
	}

	client.conns = grpcConnections

	return client, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error
	for _, conn := range c.conns {
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (c *Client) GetConn() *grpc.ClientConn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	idx := atomic.AddUint64(&c.index, 1) % uint64(len(c.conns))
	return c.conns[idx]
}

func (c *Client) Reconnect(oldConn *grpc.ClientConn) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, conn := range c.conns {
		if conn == oldConn {
			newConn, err := grpc.NewClient(c.addr, c.dialOptions...)
			if err != nil {
				return err
			}
			_ = conn.Close()
			c.conns[i] = newConn
			return nil
		}
	}
	return nil
}
