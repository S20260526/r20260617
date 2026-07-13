FROM golang:alpine AS b
WORKDIR /b

COPY internal/ /b/internal/
COPY entry/ /b/entry/
COPY hcheck/ /b/hcheck
COPY go.work /b/

RUN go mod download
RUN go build ./entry
RUN go build ./hcheck/

FROM scratch
COPY --from=b /b/m20260618-entry /m20260618-entry
COPY --from=b /b/m20260618-hcheck /m20260618-hcheck
ENTRYPOINT ["/m20260618-entry"]
HEALTHCHECK --timeout=5s --interval=10s --start-period=10s CMD ["/m20260618-hcheck"]
EXPOSE 8080
