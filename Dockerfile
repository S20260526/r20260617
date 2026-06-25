FROM golang:alpine AS b
WORKDIR /b

COPY internal/ /b/internal/
COPY cmd/ /b/cmd/
COPY hcheck/ /b/hcheck
COPY go.work /b/

RUN go mod download
RUN go build ./cmd
RUN go build ./hcheck/

FROM scratch
COPY --from=b /b/m20260618 /m20260618
COPY --from=b /b/m20260618-hcheck /m20260618-hcheck
ENTRYPOINT ["/m20260618"]
EXPOSE 8080
