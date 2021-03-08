# grpc-go-server-example

Example of grpc service.

## INSTALL
The cert folder should contain the x509 certificate: certificate.crt and privateKey.key 
The certificate can be generated with the following command openssl 
```openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -keyout privateKey.key -out certificate.crt```

Service can be deployed with using docker compose
```docker-compose up```

The following environment variables can be used in order to specify the machine where memcached is located\
MEMCACHE_HOST\
MEMCACHE_PORT\

The test client can be run from root directory with using follwing command:
```go run client\main.go```

Or sent the POST request to the localhost:50051/v1/fibonacciSlice with json object:
Example:
```{"start":0,"end":0}```

The integration Unit Tests can run with using following command:
```docker-compose -f .\docker-compose.test.yml -p ci up --abort-on-container-exit```

## Note
The int64 is using in service in order to support bigger value need to change the type ([]byte..)