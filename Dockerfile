############# builder
FROM golang:1.24.4 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-provider-equinix-metal

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG EFFECTIVE_VERSION

RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# gardener-extension-provider-equinix-metal
FROM gcr.io/distroless/static-debian11:nonroot AS gardener-extension-provider-equinix-metal
WORKDIR /

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-equinix-metal /gardener-extension-provider-equinix-metal
ENTRYPOINT ["/gardener-extension-provider-equinix-metal"]
