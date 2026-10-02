FROM debian:trixie AS p

WORKDIR /b

RUN apt-get update
RUN apt-get install -y protoc-gen-go-grpc protoc-gen-go
RUN apt-get install -y python3-grpcio python3-grpc-tools

COPY internal/grpcipc/proto/ /b/proto/

RUN mkdir golang python

RUN python3 -m grpc_tools.protoc \
        --go_out=golang --go-grpc_out=golang \
        --python_out=python --grpc_python_out=python \
        --proto_path=proto/ \
        ipc.proto

FROM golang:alpine AS b

WORKDIR /b

COPY . /b/
COPY --from=p /b/golang/internal/grpcipc/ /b/internal/grpcipc/

RUN go build ./cmd/front
RUN go build ./cmd/worker
RUN go build ./cmd/janitor

FROM debian:trixie

RUN apt-get update
RUN apt-get install -y python3-grpcio

COPY --from=p /b/python/*.py /opt/python/
COPY --from=b /b/front /b/worker /b/janitor /opt/
