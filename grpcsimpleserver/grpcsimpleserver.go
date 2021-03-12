package grpcsimpleserver

import (
	"log"
	"os"

	"encoding/binary"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"

	pb "grpc-go-server-example/grpcserver"
)

var lastIndexKey = "Last"
var lastNegIndexKey = "LastNeg"

const (
	defaultMemcasheHost  = "localhost"
	defaultMemcashePort  = "11211"
	memcachevariableHost = "MEMCACHE_HOST"
	memcachevariablePort = "MEMCACHE_PORT"
)

type fibonacciContainer interface {
	GetFibonacciValue(index int64) int64
}

type fibonacciMemcache struct {
}

func (storage fibonacciMemcache) GetMemcacheAdress() string {
	address := os.Getenv(memcachevariableHost)
	port := os.Getenv(memcachevariablePort)

	if address == "" {
		address = defaultMemcasheHost
	}
	if port == "" {
		port = defaultMemcashePort
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
	if err = storage.SaveValue(lastIndexKey, index+1); err != nil {
		log.Fatalf("%v", err)
	}

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
	if err = storage.SaveValue(lastNegIndexKey, index-1); err != nil {
		log.Fatalf("%v", err)
	}

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

// Server is the base struct represents the service and the storage that is used to cach values
type Server struct {
	pb.UnimplementedFibonacciServiceServer
	storage fibonacciContainer
}

// CreateServer create server with memcache storage
func CreateServer() Server {
	return Server{storage: fibonacciMemcache{}}
}

// FibonacciSlice is the api action
func (s *Server) FibonacciSlice(in *pb.FibonacciRequest, srv pb.FibonacciService_FibonacciSliceServer) error {
	log.Printf("Received: %v", in.Start)
	for i := in.Start; i <= in.End; i++ {
		resp := pb.Item{Value: s.storage.GetFibonacciValue(i)}
		if err := srv.Send(&resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
	return nil
}
