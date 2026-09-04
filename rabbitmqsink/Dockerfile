FROM golang:alpine AS b
WORKDIR /b

COPY go.work /b/
COPY entry/ /b/entry/
COPY hcheck/ /b/hcheck/
COPY hello/ /b/hello/
COPY world/ /b/world/
COPY internal/ /b/internal/
COPY rabbitmqsink/ /b/rabbitmqsink/

RUN go build ./rabbitmqsink

FROM scratch

COPY --from=b /b/m20260618-rabbitmqsink /m20260618-rabbitmqsink

ENTRYPOINT ["/m20260618-rabbitmqsink"]
