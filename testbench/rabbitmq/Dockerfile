FROM golang:alpine AS b
WORKDIR /b

COPY go.work /b/
COPY internal/ /b/internal/
COPY testbench/rabbitmq /b/testbench/rabbitmq/

RUN go build /b/testbench/rabbitmq

FROM scratch

COPY --from=b /b/rabbitmq /entry

ENTRYPOINT ["/entry"]
