FROM golang:alpine AS b
WORKDIR /b

COPY go.work /b/
COPY entry/ /b/entry/
COPY hcheck/ /b/hcheck/
COPY hello/ /b/hello/
COPY world/ /b/world/
COPY internal/ /b/internal/

RUN go build ./hello

FROM ubuntu:latest AS s

RUN apt-get update
RUN apt-get install -y openssl

COPY ca.crt/ /
COPY ca.key/ /

RUN openssl genrsa -out server.key
RUN openssl req -batch -new -key server.key -out server.csr -subj '/CN=hello'
RUN openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -out server.crt -days 3650

FROM scratch

COPY --from=b /b/m20260618-hello /m20260618-hello

COPY --from=s /ca.crt /
COPY --from=s /server.crt /
COPY --from=s /server.key /

ENTRYPOINT ["/m20260618-hello"]
