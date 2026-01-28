FROM golang:1.24

COPY ./go-libs /src

RUN cd /src && go mod tidy && go mod vendor && go test -v -cover ./...
