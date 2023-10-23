#
# dezswap-api-service
#
# build:
#   docker build --force-rm -t dezswap/dezswap-api-service .
# run:
#   docker run --rm -it --env-file=path/to/.env --name dezswap-api-app dezswap/dezswap-api-service

### BUILD
FROM golang:1.20-alpine AS build
ARG APP_TYPE=indexer
ARG LIBWASMVM_VERSION=v1.0.0

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
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 7d2239e9f25e96d0d4daba982ce92367aacf0cbd95d2facb8442268f2b1cc1fc
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep f6282df732a13dec836cda1f399dd874b1e3163504dbd9607c6af915b2740479
RUN cp /lib/libwasmvm_muslc.`uname -m`.a /lib/libwasmvm_muslc.a

# install simapp, remove packages
RUN go build -mod=readonly -tags "netgo muslc" \
            -ldflags "-X github.com/cosmos/cosmos-sdk/version.BuildTags='netgo,muslc' \
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
