package grpcserver

import (
	"context"
	"log"
	"net"

	"github.com/NichiNect/cachedist/internal/cache"
	"github.com/NichiNect/cachedist/pkg/pb"
	"google.golang.org/grpc"
)

type CacheServer struct {
	pb.UnimplementedCacheServiceServer
	cache cache.Cache
}

func NewCacheServer(c cache.Cache) *CacheServer {
	return &CacheServer{
		cache: c,
	}
}

func (s *CacheServer) Replicate(ctx context.Context, req *pb.ReplicateRequest) (*pb.ReplicateResponse, error) {
	if req.Operation == pb.Operation_SET {
		s.cache.Set(req.Key, req.Value, int(req.TtlSeconds))
	} else if req.Operation == pb.Operation_DELETE {
		s.cache.Delete(req.Key)
	}

	return &pb.ReplicateResponse{
		Ok:    true,
		Error: "",
	}, nil
}

func StartServer(port string, c cache.Cache) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCacheServiceServer(grpcServer, NewCacheServer(c))

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
