FROM golang:alpine AS b
WORKDIR /b

COPY go.work /b/
COPY entry/ /b/entry/
COPY hcheck/ /b/hcheck/
COPY hello/ /b/hello/
COPY world/ /b/world/
COPY internal/ /b/internal/

RUN go build ./world

FROM ubuntu:latest AS c

RUN apt-get update
RUN apt-get install -y openssl

COPY world/openssl.cnf /

RUN openssl genrsa -out server.key 4096
RUN openssl req -new -key server.key -out server.csr -config openssl.cnf
RUN --mount=type=secret,id=CA_KEY,target=/ca.key --mount=type=secret,id=CA_CRT,target=/ca.crt openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -out server.crt -days 3650 -extfile openssl.cnf -extensions v3_req
RUN --mount=type=secret,id=CA_CRT,target=/ca.crt.src cp /ca.crt.src /ca.crt

FROM scratch

COPY --from=b /b/m20260618-world /m20260618-world

COPY --from=c /ca.crt /
COPY --from=c /server.crt /
COPY --from=c /server.key /

ENTRYPOINT ["/m20260618-world"]
