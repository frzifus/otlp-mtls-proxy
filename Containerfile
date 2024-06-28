FROM golang:1.22

COPY . /go/src/github.com/frzifus/otel-mtls-proxy

WORKDIR /go/src/github.com/frzifus/otel-mtls-proxy

RUN CGO_ENABLED=0 go build -v -o /otel-mtls-proxy *.go

FROM scratch

COPY --from=0 /otel-mtls-proxy /otel-mtls-proxy

EXPOSE 4318

CMD ["/otel-mtls-proxy"]
