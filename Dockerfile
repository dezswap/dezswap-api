#
# dezswap-api-service
#
# build:
#   docker build --force-rm -t dezswap/dezswap-api-service .
# run:
#   docker run --rm -it --env-file=path/to/.env --name dezswap-api-app dezswap/dezswap-api-service

### BUILD
FROM golang:1.23-alpine AS build
ARG APP_TYPE=indexer
ARG LIBWASMVM_VERSION=v2.2.4
ARG APP_VERSION=dev

WORKDIR /app

# Create appuser.
RUN adduser -D -g '' appuser
# Install required binaries
RUN apk add --update --no-cache git build-base linux-headers

# Copy app dependencies
COPY go.mod go.mod
COPY go.sum go.sum
COPY Makefile Makefile
# Download all golang package dependencies
RUN make deps

# Copy source files
COPY . .

# See https://github.com/CosmWasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/${LIBWASMVM_VERSION}/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/${LIBWASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 27fb13821dbc519119f4f98c30a42cb32429b111b0fdc883686c34a41777488f
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep 70c989684d2b48ca17bbd55bb694bbb136d75c393c067ef3bdbca31d2b23b578
RUN cp /lib/libwasmvm_muslc.`uname -m`.a /lib/libwasmvm_muslc.a

# install simapp, remove packages
RUN go build -mod=readonly -tags "netgo muslc" \
            -ldflags "\
            -X github.com/cosmos/cosmos-sdk/version.BuildTags='netgo,muslc' \
            -X github.com/dezswap/dezswap-api/api.AppVersion=${APP_VERSION} \
            -w -s -linkmode=external -extldflags '-Wl,-z,muldefs -static'" \
            -trimpath -o ./main ./cmd/${APP_TYPE}

### RELEASE
FROM alpine:latest AS release
WORKDIR /app
# Import the user and group files to run the app as an unpriviledged user
COPY --from=build /etc/passwd /etc/passwd

COPY --from=build /app/config.yml /app/config.yml

# Use an unprivileged user
USER appuser
COPY --from=build /app/cmd /app/cmd
# Grab compiled binary from build
COPY --from=build /app/main /app/main

# Set entry point
CMD [ "./main" ]
