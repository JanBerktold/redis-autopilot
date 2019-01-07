FROM golang:1.11.4-stretch AS builder
WORKDIR /target

COPY src/ ./

RUN CGO_ENABLED=0 && GOOS=linux && GOARCH=amd64 && GO111MODULE=on && go build -o pilot

FROM redis:5.0.3-stretch

ENV PYTHON_VERSION=2.7.12-r1
ENV PY_PIP_VERSION=8.1.2-r0
ENV SUPERVISOR_VERSION=3.3.1

RUN apt-get update && apt-get install -y --no-install-recommends python python-pip python-setuptools
RUN pip install supervisor==$SUPERVISOR_VERSION

COPY --from=builder /target/pilot /usr/local/bin/autopilot
RUN chmod +x /usr/local/bin/autopilot
COPY supervisord.conf /etc/supervisord.conf
COPY default_config.yml /etc/redis-autopilot/config

ENTRYPOINT ["supervisord"]
