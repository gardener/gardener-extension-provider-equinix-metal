############# builder
FROM golang:1.13.4 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-provider-packet
COPY . .
RUN make install-requirements && make VERIFY=true all

############# gardener-extension-provider-packet
FROM alpine:3.11.3 AS gardener-extension-provider-packet

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-packet /gardener-extension-provider-packet
ENTRYPOINT ["/gardener-extension-provider-packet"]
