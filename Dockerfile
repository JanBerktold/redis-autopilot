FROM golang:1.11.4-stretch AS builder
WORKDIR /target

# Get the go-enum package first to allow docker to cache it regardless of our code
RUN go get github.com/abice/go-enum

# We want to populate the module cache based on the go.{mod,sum} files.
COPY src/go.mod .
COPY src/go.sum .

# Download the modules to allow docker to cache this step.
RUN go mod download

COPY src/ ./

RUN CGO_ENABLED=0 \
    && GOOS=linux  \
    && GOARCH=amd64 \
    && GO111MODULE=on \
    && go generate \
    && go build -o pilot

FROM redis:5.0.3-stretch

ENV PYTHON_VERSION=2.7.12-r1
ENV PY_PIP_VERSION=8.1.2-r0
ENV SUPERVISOR_VERSION=3.3.1

RUN apt-get update \
    && apt-get install -y --no-install-recommends python python-pip python-setuptools \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && pip install supervisor==$SUPERVISOR_VERSION

COPY --from=builder /target/pilot /usr/local/bin/redis-autopilot
RUN chmod +x /usr/local/bin/redis-autopilot
COPY supervisord.conf /etc/supervisord.conf
COPY redis.conf /etc/redis.conf
COPY default_config.yml /etc/redis-autopilot/config.yml

ENTRYPOINT ["supervisord", "-c", "/etc/supervisord.conf"]
