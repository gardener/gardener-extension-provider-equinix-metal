############# builder
FROM eu.gcr.io/gardener-project/3rd/golang:1.15.5 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-provider-packet
COPY . .
RUN make install

############# gardener-extension-provider-packet
FROM eu.gcr.io/gardener-project/3rd/alpine:3.12.1 AS gardener-extension-provider-packet

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-packet /gardener-extension-provider-packet
ENTRYPOINT ["/gardener-extension-provider-packet"]
