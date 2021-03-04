# grpc-go-server-example

## INSTALL

Service can be deployed with using docker compose
```docker-compose up```


The following environment variables can be used in order to specify the machine where memcached is located\
MEMCACHE_HOST\
MEMCACHE_PORT\

The cert folder should contain the x509 certificate: certificate.crt and privateKey.key 
the certificate can be generated with using openssl 
```req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -keyout privateKey.key -out certificate.crt```

The test client can be run from root directory with using follwing command:
```go run client\main.go```