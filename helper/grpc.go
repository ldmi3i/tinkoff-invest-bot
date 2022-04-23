package helper

import (
	"crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
)

var cl *grpc.ClientConn

func initGRPC() {
	addr := GetGRPCAddress()
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
	return cl.Close()
}
