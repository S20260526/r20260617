FROM golang:alpine AS b
WORKDIR /b

COPY go.work /b/
COPY entry/ /b/entry/
COPY hcheck/ /b/hcheck/
COPY hello/ /b/hello/
COPY world/ /b/world/
COPY internal/ /b/internal/

RUN go build ./world

FROM scratch

COPY --from=b /b/m20260618-world /m20260618-world

ENTRYPOINT ["/m20260618-world"]
