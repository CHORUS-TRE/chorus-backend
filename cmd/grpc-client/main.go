package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"google.golang.org/grpc"
)

func main() {
	address := flag.String("server", "", "gRPC server in format host:port")
	flag.Parse()

	//nolint: staticcheck
	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := chorus.NewAuthenticationServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &chorus.Credentials{
		Username: "chorus-admin",
		Password: "superpassword",
	}
	//nolint: staticcheck
	res, err := client.Authenticate(ctx, req)
	if err != nil {
		log.Fatalf("authentication failed: %v", err)
	}
	log.Printf("authentication result: %v\n", res.Result.Token)
}
