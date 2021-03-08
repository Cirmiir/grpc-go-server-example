package main

import (
	"context"
	pb "grpc-go-server-example/grpcserver"
	"io"
	"log"
	"net"
	"testing"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

type fibonacciStorage struct {
	Storage []int64
}

func (storage fibonacciStorage) GetFibonacciValue(index int64) int64 {
	if index < 1 {
		return 0
	}

	switch index {
	case 1:
		return 1
	}

	if int(index) >= len(storage.Storage) {
		for i := len(storage.Storage); i <= int(index); i++ {
			storage.Storage = append(storage.Storage, storage.Storage[i-2]+storage.Storage[i-1])
		}
		return storage.Storage[index]
	}
	return storage.Storage[index]
}

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	var initValue = []int64{0, 1, 1}
	pb.RegisterFibonacciServiceServer(s, &server{storage: fibonacciStorage{Storage: initValue}})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func prepareTestCase(start int64, end int64, expectedResult []int64) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
		if err != nil {
			t.Fatalf("Failed to dial bufnet: %v", err)
		}
		defer conn.Close()
		client := pb.NewFibonacciServiceClient(conn)

		stream, err := client.FibonacciSlice(ctx, &pb.FibonacciRequest{Start: start, End: end})
		if err != nil {
			t.Fatalf("could not greet: %v", err)
		}
		done := make(chan bool)
		actualResult := make([]int64, 0)

		go func() {
			index := 0
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					done <- true //means stream is finished
					return
				}
				if err != nil {
					log.Fatalf("cannot receive %v", err)
				}
				t.Logf("Resp received: %v", resp.Value)
				actualResult = append(actualResult, resp.Value)
				index = index + 1
			}
		}()

		<-done //we will wait until all response is received

		if len(expectedResult) != len(actualResult) {
			t.Fatalf("The lengths of expected and actual results are different. Length expected: %v, actual: %v ", len(expectedResult), len(actualResult))
		}

		for i, v := range actualResult {
			if v != expectedResult[i] {
				t.Fatalf("The value at index %v is mismatched. Expected: %v Actual: %v", i, expectedResult[i], v)
			}
		}
	}
}

func TestFibonacciSimple(t *testing.T) {
	t.Run("test 1", prepareTestCase(1, 2, []int64{1, 1}))
	t.Run("test 2", prepareTestCase(6, 7, []int64{8, 13}))
	t.Run("test 3", prepareTestCase(6, 3, []int64{}))
	t.Run("test 3", prepareTestCase(3, 10, []int64{2, 3, 5, 8, 13, 21, 34, 55}))
}
