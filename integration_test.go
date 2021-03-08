package main

import (
	pb "grpc-go-server-example/grpcserver"
	"log"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var storage = fibonacciMemcache{}

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterFibonacciServiceServer(s, &server{storage: storage})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func TestFibonacciIntegration(t *testing.T) {
	t.Run("test 1", prepareTestCase(1, 2, []int64{1, 1}))
	t.Run("test 2", prepareTestCase(6, 7, []int64{8, 13}))
	t.Run("test 3", prepareTestCase(-3, -1, []int64{2, -1, 1}))
	t.Run("test 3", prepareTestCase(6, 3, []int64{}))
}

func TestLastIndex(t *testing.T) {
	mc := memcache.New(storage.GetMemcacheAdress())
	err := mc.FlushAll()

	if err != nil {
		t.Fatalf("Can not perform flush for memcache")
	}

	_, err = storage.GetValue(lastIndexKey)

	if err != memcache.ErrCacheMiss {
		t.Fatalf("Index should not be set by default")
	}

	prepareTestCase(6, 7, []int64{8, 13})(t)

	index, err := storage.GetValue(lastIndexKey)

	if err == memcache.ErrCacheMiss {
		t.Fatalf("Index wasn't save after get fibonacci request.")
	}

	if err != nil {
		t.Fatalf("The last index cannot be retrieved.")
	}

	if index != 8 {
		t.Fatalf("The last index was not updated properly. Expected: 7 Actual: %v", index)
	}
}
