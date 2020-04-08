############# builder
FROM golang:1.13.8 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-provider-packet
COPY . .
RUN make install

############# gardener-extension-provider-packet
FROM alpine:3.11.3 AS gardener-extension-provider-packet

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-packet /gardener-extension-provider-packet
ENTRYPOINT ["/gardener-extension-provider-packet"]
