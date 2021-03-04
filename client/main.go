package main

import (
	"crypto/tls"
	"io"
	"log"
	"time"

	cert "grpc-go-server-example/cert"
	pb "grpc-go-server-example/grpcserver"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	dcreds := credentials.NewTLS(&tls.Config{
		ServerName:         "localhost:50051",
		RootCAs:            cert.DemoCertPool,
		InsecureSkipVerify: true,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}
	conn, err := grpc.Dial(address, dopts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewFibonacciServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := c.FibonacciSlice(ctx, &pb.FibonacciRequest{Start: 6, End: 7})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	done := make(chan bool)

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- true //means stream is finished
				return
			}
			if err != nil {
				log.Fatalf("cannot receive %v", err)
			}
			log.Printf("Resp received: %v", resp.Value)
		}
	}()

	<-done //we will wait until all response is received
	log.Printf("finished")
}
