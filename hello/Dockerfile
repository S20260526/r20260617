FROM golang:alpine AS b
WORKDIR /b

COPY go.work /b/
COPY entry/ /b/entry/
COPY hcheck/ /b/hcheck/
COPY hello/ /b/hello/
COPY world/ /b/world/
COPY internal/ /b/internal/

RUN go build ./hello

FROM scratch
COPY --from=b /b/m20260618-hello /m20260618-hello
ENTRYPOINT ["/m20260618-hello"]
