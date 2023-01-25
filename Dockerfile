ARG BUILDPLATFORM=linux/amd64
FROM --platform=$BUILDPLATFORM golang:1.18 AS build-env
RUN mkdir -p /go/src/delegated-client
WORKDIR /go/src/delegated-client
COPY  . .
ARG TARGETARCH
ARG TAG
ARG COMMIT
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH GOOS=linux go build -a -o dclient \
      -ldflags "-s -w -X main.VERSION=$TAG -X main.COMMIT=$COMMIT" ./cmd/dclient

RUN CGO_ENABLED=0 GOARCH=$TARGETARCH GOOS=linux go build -a -o sclient \
      -ldflags "-s -w -X main.VERSION=$TAG -X main.COMMIT=$COMMIT" ./cmd/sclient

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build-env /go/src/delegated-client/dclient .
COPY --from=build-env /go/src/delegated-client/sclient .
# Use uid of nonroot user (65532) because kubernetes expects numeric user when applying pod security policies
USER 65532
ENTRYPOINT ["/dclient"]
