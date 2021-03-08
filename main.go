package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"encoding/binary"
	"strconv"

	"context"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	cert "grpc-go-server-example/cert"
	pb "grpc-go-server-example/grpcserver"

	grpc "google.golang.org/grpc"

	"google.golang.org/grpc/credentials"
)

const (
	host = "localhost"
	port = "50051"
)

var lastIndexKey = "Last"
var lastNegIndexKey = "LastNeg"

var defaultAddress = host + ":" + port

type fibonacciContainer interface {
	GetFibonacciValue(index int64) int64
}

type fibonacciMemcache struct {
}

func (storage fibonacciMemcache) GetMemcacheAdress() string {
	address := os.Getenv("MEMCACHE_HOST")
	port := os.Getenv("MEMCACHE_PORT")

	if address == "" {
		address = "localhost"
	}
	if port == "" {
		port = "11211"
	}

	return address + ":" + port
}

func (storage fibonacciMemcache) SaveValue(key string, value int64) error {
	mc := memcache.New(storage.GetMemcacheAdress())
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(value))
	err := mc.Set(&memcache.Item{Key: key, Value: bs})
	return err
}

func (storage fibonacciMemcache) GetValue(key string) (int64, error) {
	mc := memcache.New(storage.GetMemcacheAdress())
	it, err := mc.Get(key)

	if err == nil {
		return int64(binary.LittleEndian.Uint64(it.Value)), err
	}
	return 0, err
}

func (storage fibonacciMemcache) init() error {
	err := storage.SaveValue(strconv.Itoa(0), 0)
	if err != nil {
		return err
	}
	err = storage.SaveValue(strconv.Itoa(1), 1)

	if err != nil {
		return err
	}

	return nil
}

func (storage fibonacciMemcache) GetNextPositive(index int64) int64 {
	var startIndex, prev, current int64
	var err error
	prev = 0
	current = 1
	if startIndex, err = storage.GetValue(lastIndexKey); err == nil {
		current, _ = storage.GetValue(strconv.Itoa(int(startIndex - 1)))
		prev, _ = storage.GetValue(strconv.Itoa(int(startIndex - 2)))
	} else if err == memcache.ErrCacheMiss {
		startIndex = 2
		if err = storage.init(); err != nil {
			log.Fatalf("%v", err)
		}
	} else {
		log.Fatalf("%v", err)
		return 0
	}

	for i := startIndex; i <= index; i++ {
		current, prev = current+prev, current
		if err = storage.SaveValue(strconv.Itoa(int(i)), current); err != nil {
			log.Fatalf("%v", err)
		}
	}
	storage.SaveValue(lastIndexKey, index+1)

	return current
}
func (storage fibonacciMemcache) GetNextNegative(index int64) int64 {
	var startIndex, prev, current int64
	var err error
	prev = 1
	current = 0
	if startIndex, err = storage.GetValue(lastNegIndexKey); err == nil {
		current, _ = storage.GetValue(strconv.Itoa(int(startIndex + 1)))
		prev, _ = storage.GetValue(strconv.Itoa(int(startIndex + 2)))
	} else if err == memcache.ErrCacheMiss {
		startIndex = -1
		if err = storage.init(); err != nil {
			log.Fatalf("%v", err)
		}
	} else {
		log.Fatalf("%v", err)
		return 0
	}

	for i := startIndex; i >= index; i-- {
		current, prev = prev-current, current
		if err = storage.SaveValue(strconv.Itoa(int(i)), current); err != nil {
			log.Fatalf("%v", err)
		}
	}
	storage.SaveValue(lastNegIndexKey, index-1)

	return current
}

func (storage fibonacciMemcache) GetFibonacciValue(index int64) int64 {
	it, err := storage.GetValue(strconv.Itoa(int(index)))

	if err == nil {
		return it
	}

	switch index {
	case 0:
		return 0
	case 1:
		return 1
	}

	if err == memcache.ErrCacheMiss {
		if index >= 0 {
			return storage.GetNextPositive(index)
		}
		return storage.GetNextNegative(index)
	}

	if err != nil {
		log.Fatalf("%v", err)
	}

	return 0

}

type server struct {
	pb.UnimplementedFibonacciServiceServer
	storage fibonacciContainer
}

func (s *server) FibonacciSlice(in *pb.FibonacciRequest, srv pb.FibonacciService_FibonacciSliceServer) error {
	log.Printf("Received: %v", in.Start)
	for i := in.Start; i <= in.End; i++ {
		resp := pb.Item{Value: s.storage.GetFibonacciValue(i)}
		if err := srv.Send(&resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
	return nil
}

func main() {
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewClientTLSFromCert(cert.DemoCertPool, defaultAddress))}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterFibonacciServiceServer(grpcServer, &server{storage: fibonacciMemcache{}})
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
