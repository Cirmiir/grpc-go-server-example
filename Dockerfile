FROM golang:1.14.4

WORKDIR /go/src/grpc-go-server-example

# Copy the local package files to the container's workspace.
COPY . .
RUN go mod init
RUN go build .


# Document that the service listens on port 8080.
EXPOSE 50051

CMD ["go", "run", "-mod=mod" , "."]