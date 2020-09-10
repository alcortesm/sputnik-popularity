ARG project=sputnik-popularity
ARG src=/tmp/${project}

####
# Creates a Docker image with the sources.
FROM golang:1.15.1-buster AS src
ARG src

COPY . ${src}
WORKDIR ${src}

####
# Creates a Docker image with the static binary to run the app.
FROM src AS build-app

RUN CGO_ENABLED=0 go build \
    -o /bin/build \
    -ldflags '-extldflags "-static"' \
    -tags timetzdata \
    ./app/cmd/sputnik-popularity

####
# Creates a Docker image with the scrapme helper binary for the e2e tests.
FROM src AS build-scrapeme

RUN CGO_ENABLED=0 go build \
    -o /bin/build \
    -ldflags '-extldflags "-static"' \
    ./tests/e2e/scrapeme

####
# Creates an empty image with certs and other things required
# to run Go static binaries.
FROM scratch AS with-certs

COPY --from=src \
    /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt

####
# Creates a single layer image to run the app.
FROM with-certs AS run-app

COPY --from=build-app /bin/build /bin/runme
ENTRYPOINT ["/bin/runme"]

####
# Creates a single layer image to run the scrapeme helper for the e2e
# tests.
FROM scratch AS run-scrapeme

COPY --from=build-scrapeme /bin/build /bin/runme
ENTRYPOINT ["/bin/runme"]
