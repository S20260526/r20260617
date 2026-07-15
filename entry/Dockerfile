FROM golang:alpine AS b
WORKDIR /b

COPY go.work /b/
COPY entry/ /b/entry/
COPY hcheck/ /b/hcheck/
COPY hello/ /b/hello/
COPY world/ /b/world/
COPY internal/ /b/internal/

RUN go build ./entry
RUN go build ./hcheck/

FROM ubuntu:latest AS c

RUN apt-get update
RUN apt-get install -y openssl

COPY ca.crt/ /
COPY ca.key/ /
COPY entry/openssl.cnf /

RUN openssl genrsa -out client.key 4096
RUN openssl req -new -key client.key -out client.csr -config openssl.cnf
RUN openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -out client.crt -days 3650 -extfile openssl.cnf -extensions v3_req

FROM scratch
COPY --from=b /b/m20260618-entry /m20260618-entry
COPY --from=b /b/m20260618-hcheck /m20260618-hcheck
ENTRYPOINT ["/m20260618-entry"]
HEALTHCHECK --timeout=5s --interval=10s --start-period=10s CMD ["/m20260618-hcheck"]
EXPOSE 8080
