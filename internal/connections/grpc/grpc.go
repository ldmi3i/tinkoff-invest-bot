package grpc

import (
	"crypto/tls"
	"github.com/ldmi3i/tinkoff-invest-bot/internal/env"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
)

var cl *grpc.ClientConn

func InitGRPC() {
	addr := env.GetGRPCAddress()
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	if err != nil {
		log.Fatalf("Fatal error while connection grpc:\n%s", err)
	}
	cl = conn
}

func GetClient() grpc.ClientConnInterface {
	return cl
}

func Close() error {
	log.Println("Gracefully closing grpc connection")
	if cl == nil {
		return nil
	}
	return cl.Close()
}
