FROM golang:alpine AS b
WORKDIR /b

COPY . /b/

RUN rm internal/infra/grpcipc.go
RUN go build ./testbench/rabbitmq

FROM scratch

COPY --from=b /b/rabbitmq /entry

ENTRYPOINT ["/entry"]
