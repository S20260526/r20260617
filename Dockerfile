FROM golang:alpine AS b
WORKDIR /b

COPY internal/ /b/internal/
COPY cmd/ /b/cmd/
COPY go.work /b/

RUN go mod download
RUN go build ./cmd

FROM scratch
COPY --from=b /b/m20260618 /m20260618
ENTRYPOINT ["/m20260618"]
EXPOSE 8080
