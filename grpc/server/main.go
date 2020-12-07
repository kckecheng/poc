package main

import (
	"context"
	"log"
	"net"

	pb "github.com/kckecheng/poc/grpc/snode"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type snodeServer struct {
	pb.UnimplementedSNodeServer
}

func (s *snodeServer) Execute(ctx context.Context, cmd *pb.Command) (*pb.Result, error) {
	log.Printf("Received: %v", cmd.GetCommand())
	return &pb.Result{Output: "TBD"}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSNodeServer(s, &snodeServer{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
