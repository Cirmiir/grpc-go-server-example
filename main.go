package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	cert "grpc-go-server-example/cert"
	pb "grpc-go-server-example/grpcserver"
	server "grpc-go-server-example/grpcsimpleserver"

	grpc "google.golang.org/grpc"

	"google.golang.org/grpc/credentials"
)

const (
	host = "localhost"
	port = "50051"
)

var defaultAddress = host + ":" + port

func main() {

	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewClientTLSFromCert(cert.DemoCertPool, defaultAddress))}

	grpcServer := grpc.NewServer(opts...)
	fibonacciService := server.CreateServer()
	pb.RegisterFibonacciServiceServer(grpcServer, &fibonacciService)
	ctx := context.Background()

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName:         defaultAddress,
		RootCAs:            cert.DemoCertPool,
		InsecureSkipVerify: true,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	mux := http.NewServeMux()

	gwmux := runtime.NewServeMux()
	err := pb.RegisterFibonacciServiceHandlerFromEndpoint(ctx, gwmux, defaultAddress, dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return
	}

	mux.Handle("/", gwmux)

	conn, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:    defaultAddress,
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert.DemoKeyPair},
			NextProtos:   []string{"h2"},
		},
	}

	fmt.Printf("grpc on port: %v\n", port)
	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}
