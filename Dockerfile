############# builder
FROM golang:1.17.5 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-provider-equinix-metal
COPY . .
RUN make install

############# gardener-extension-provider-equinix-metal
FROM gcr.io/distroless/static-debian11:nonroot AS gardener-extension-provider-equinix-metal
WORKDIR /

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-equinix-metal /gardener-extension-provider-equinix-metal
ENTRYPOINT ["/gardener-extension-provider-equinix-metal"]
