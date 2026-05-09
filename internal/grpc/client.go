package grpcserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/NichiNect/cachedist/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
}

func NewGRPCClient() *GRPCClient {
	return &GRPCClient{
		conns: make(map[string]*grpc.ClientConn),
	}
}

func (c *GRPCClient) getConn(addr string) (*grpc.ClientConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if conn, ok := c.conns[addr]; ok {
		return conn, nil
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c.conns[addr] = conn
	return conn, nil
}

func (c *GRPCClient) Replicate(addr string, req *pb.ReplicateRequest) error {
	conn, err := c.getConn(addr)
	if err != nil {
		return fmt.Errorf("failed to get connection for %s: %w", addr, err)
	}

	client := pb.NewCacheServiceClient(conn)
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.Replicate(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC Replicate failed: %w", err)
	}

	if !resp.Ok {
		return fmt.Errorf("replication returned error: %s", resp.Error)
	}

	return nil
}
