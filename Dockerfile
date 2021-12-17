############# builder
FROM eu.gcr.io/gardener-project/3rd/golang:1.17.5 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-provider-equinix-metal
COPY . .
RUN make install

############# gardener-extension-provider-equinix-metal
FROM eu.gcr.io/gardener-project/3rd/alpine:3.15.0 AS gardener-extension-provider-equinix-metal

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-equinix-metal /gardener-extension-provider-equinix-metal
ENTRYPOINT ["/gardener-extension-provider-equinix-metal"]
