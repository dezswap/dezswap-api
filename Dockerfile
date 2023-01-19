#
# dezswap-api-service
#
# build:
#   docker build --force-rm -t dezswap/dezswap-api-service .
# run:
#   docker run --rm -it --env-file=path/to/.env --name dezswap-api-app dezswap/dezswap-api-service

### BUILD
FROM golang:1.19-alpine AS build
ARG APP_TYPE=indexer

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

# install simapp, remove packages
RUN go build -mod=readonly -tags "netgo muslc" -w -s -o ./main ./cmd/${APP_TYPE}

### RELEASE
FROM alpine:latest AS release
WORKDIR /app
# Import the user and group files to run the app as an unpriviledged user
COPY --from=build /etc/passwd /etc/passwd

COPY --from=build /app/config.yaml /app/config.yaml

# Use an unprivileged user
USER appuser
COPY --from=build /app/cmd /app/cmd
# Grab compiled binary from build
COPY --from=build /app/main /app/main

# Set entry point
CMD [ "./main" ]
