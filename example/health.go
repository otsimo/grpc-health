package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	mut     sync.Mutex
	counter int = 0
)

type healthServer struct{}

func (hs *healthServer) Check(context.Context, *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	mut.Lock()
	defer mut.Unlock()
	counter++
	log.Println("counter=", counter)
	switch counter % 4 {
	case 0:
		return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
	case 1:
		return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_NOT_SERVING}, nil
	case 2:
		return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_UNKNOWN}, nil
	default:
		return nil, fmt.Errorf("it is 3")
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	healthpb.RegisterHealthServer(s, &healthServer{})
	log.Println("staring server")
	s.Serve(lis)
}
