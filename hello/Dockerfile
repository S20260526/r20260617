FROM golang:alpine AS b
WORKDIR /b

COPY internal/ /b/internal/
COPY hello/ /b/hello/
COPY go.work /b/

RUN go mod download
RUN go build ./hello

FROM scratch
COPY --from=b /b/m20260618-hello /m20260618-hello
ENTRYPOINT ["/m20260618-hello"]
