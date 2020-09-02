# Creates a Docker image for running the service.
ARG name=sputnik-popularity
ARG bin=/bin/${name}
ARG src=/tmp/${name}

####
# Creates a Docker image with the sources.
FROM golang:1.15.0-buster AS src

ARG src

# Copy the sources.
COPY . ${src}

WORKDIR ${src}

####
# Creates a Docker image with a static binary of the service.
FROM src as build
ARG bin
ARG name

RUN CGO_ENABLED=0 go build \
    -o ${bin} \
    -ldflags '-extldflags "-static"' \
    ./cmd/${name}

####
# Creates a single layer image to run the service.
FROM scratch as run
ARG bin

# Copy some required files for the Go stdlib to work: the
# ca-certificates for SSL and the timezone database.  We just reuse the
# ones in the build image.
COPY --from=src \
    /usr/local/go/lib/time/zoneinfo.zip \
    /usr/local/go/lib/time/zoneinfo.zip
COPY --from=src \
    /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt

COPY --from=build ${bin} /bin/service

ENTRYPOINT ["/bin/service"]
