FROM python:trixie AS p

WORKDIR /b

RUN pip install grpcio-tools
RUN pip install grpcio
RUN apt-get update
RUN apt-get install -y protoc-gen-go-grpc protoc-gen-go

COPY internal/grpcipc/proto/ /b/proto/

RUN mkdir golang python

RUN python3 -m grpc_tools.protoc \
        --go_out=golang --go-grpc_out=golang \
        --python_out=python --grpc_python_out=python \
        --proto_path=. \
        proto/ipc.proto

FROM golang:alpine AS b

WORKDIR /b

COPY . /b/
COPY --from=p /b/golang/internal/grpcipc/ /b/internal/grpcipc/

RUN go build ./testbench/grpcipc

FROM scratch

COPY --from=b /b/grpcipc /entry

ENTRYPOINT ["/entry"]
