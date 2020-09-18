# Sputnik Popularity

[![CircleCI](https://circleci.com/gh/alcortesm/sputnik-popularity.svg?style=shield)](https://circleci.com/gh/alcortesm/sputnik-popularity)
![Lint](https://github.com/alcortesm/sputnik-popularity/workflows/Lint/badge.svg)

Sputnik populairty keeps track of how busy is my local climbing gym and
allows to see how popular it has been over the past few days.

It scrapes my gym's capacity utilization endpoint and store the data in an InfluxDB instance.
A web front end shows the utilization during the last couple of weeks.

## My Production Environment

The main restriction when building my production environemnt for this project
is that everything should be free (as in beer).

Software:

  - The Git repository is available on [Github](https://github.com/alcortesm/sputnik-popularity).

Hardware:

  - Database: an InfluxDB instance running on [Amazon Web Services](https://aws.amazon.com) (AWS) in Germany, EU.
  - Scraper and Web front-end: running as a single docker container on [Google Cloud Platform](https://cloud.google.com) (GCP) in South Carolina, USA.

Develop/Deploy process:
  - The master branch contains the version of the project currently running in production.
  - I develop all new functionalities and bug fixes as separate git branches; they get merged into master via Github pull requests.
  - I verify all tests are passing on every pull request using [CircleCI](https://circleci.com). You can do the same locally using `make test`.
  - I use Github Actions to lint every pull request using [golangci-lint](https://github.com/golangci/golangci-lint). You do the same locally using `make lint`.

## How to run the tests

There are 3 types of tests: unit, integration and e2e.
You can run all of them with `make test`.

To run them individually use `make unit`, `make integration` or `make e2e`.

## How to run the project

Some environment variables are required to run the project:

| Environment variable | Description |
|---|---|
| SPUTNIK\_INFLUXDB\_URL | your InfluxDB URL |
| SPUTNIK\_INFLUXDB\_TOKEN\_WRITE | your InfluxDB write token |
| SPUTNIK\_INFLUXDB\_TOKEN\_READ | your InfluxDB read token |
| SPUTNIK\_SCRAPE\_URL | the URL to fetch the gym's current capacity utilization |
| SPUTNIK\_SCRAPE\_GYM\_NAME | the gym name to pass to the scrape URL |
| SPUTNIK\_SCRAPE\_GYM\_ID | the gym ID to pass to the scrape URL |

Once this environment variables have been set
you can run the project locally with:

```
; go run ./app/cmd/sputnik-popularity
```

### Run as a docker container in Google Compute Engine

First build a docker image of the project:

```
; make docker-image
[...]
Successfully tagged sputnik:latest
```

This will generate an small image (~11MB) tagged `sputnik:latest` with the binary of the project.

Now, push the image to a docker container registry of your choosing.
Here I'm using the Google Cloud Container Registry
and I assume PROJECT\_ID holds the ID of one of my projects in Google Cloud:

```
; docker tag sputnik:latest gcr.io/$PROJECT_ID/sputnik:latest
; docker push gcr.io/$PROJECT_ID/sputnik:latest
```

Connect to a Google Compute Engine instance and run the docker image.
Here I assume INSTANCE\_NAME holds the name of one of my virtual machines in Google Compute Engine:

```
; gcloud compute ssh $INSTANCE_NAME
name@instance ~ $  docker run \
    --name sputnik \
    --detach \
    --env SPUTNIK_INFLUXDB_URL="..." \
    --env SPUTNIK_INFLUXDB_TOKEN_WRITE="..." \
    --env SPUTNIK_INFLUXDB_TOKEN_READ="..." \
    --env SPUTNIK_SCRAPE_URL="..." \
    --env SPUTNIK_SCRAPE_GYM_NAME="..." \
    --env SPUTNIK_SCRAPE_GYM_ID="..." \
    --publish 8080:8080 \
    --log-driver=gcplogs \
    gcr.io/$PROJECT_ID/sputnik_popularity
```

This will create a container named `sputnik` and send the app logs to the Google Cloud Logs Viewer.

Note that if your image is stored as private in your registry
you will need to configure docker in your instance to authenticate with it
before running or pulling the image.
