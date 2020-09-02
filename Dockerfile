ARG project=sputnik-popularity
ARG src=/tmp/${project}

####
# Creates a Docker image with the sources.
FROM golang:1.15.0-buster AS src
ARG src

COPY . ${src}
WORKDIR ${src}

####
# Creates a Docker image with the static binary to run the app.
FROM src as build_app

RUN CGO_ENABLED=0 go build \
    -o /bin/build \
    -ldflags '-extldflags "-static"' \
    -tags timetzdata \
    ./app/cmd/sputnik-popularity

####
# Creates a Docker image with the scrapme helper binary for the e2e tests.
FROM src as build_scrapeme

RUN CGO_ENABLED=0 go build \
    -o /bin/build \
    -ldflags '-extldflags "-static"' \
    ./tests/e2e/scrapeme

####
# Creates an empty image with certs and other things required
# to run Go static binaries.
FROM scratch as withCerts

COPY --from=src \
    /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt

####
# Creates a single layer image to run the app.
FROM withCerts as run_app

COPY --from=build_app /bin/build /bin/runme
ENTRYPOINT ["/bin/runme"]

####
# Creates a single layer image to run the scrapeme helper for the e2e
# tests.
FROM scratch as run_scrapeme

COPY --from=build_scrapeme /bin/build /bin/runme
ENTRYPOINT ["/bin/runme"]
