FROM golang:alpine AS b
WORKDIR /b

COPY app/ /b/app/
COPY cmd/ /b/cmd/
COPY domain/ /b/domain/
COPY infra/ /b/infra/
COPY go.work /b/

RUN go mod download
RUN go build ./cmd

FROM scratch
COPY --from=b /b/m20260618 /m20260618
ENTRYPOINT ["/m20260618"]
EXPOSE 8080
