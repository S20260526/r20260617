FROM golang:alpine AS b
WORKDIR /b

COPY go.work /b/
COPY entry/ /b/entry/
COPY hcheck/ /b/hcheck/
COPY hello/ /b/hello/
COPY world/ /b/world/
COPY internal/ /b/internal/

RUN go build ./hello

COPY ca.crt/ /b/
COPY ca.key/ /b/

RUN openssl genrsa -out server.key
RUN openssl req -new -key server.key -out cerver.csr
RUN openssl x509 -req -batch -in server.csr -CA ca.crt -CAkey ca.key -out server.crt -days 3650 -subj '/CN=hello'

FROM scratch

COPY --from=b /b/m20260618-hello /m20260618-hello

COPY --from=b /b/ca.crt /
COPY --from=b /b/server.crt /
COPY --from=b /b/server.key /

ENTRYPOINT ["/m20260618-hello"]
