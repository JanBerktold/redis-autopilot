# select image
FROM golang:alpine AS builder

WORKDIR /target

# copy your source tree
COPY src/ ./

# build for release
RUN go build -o pilot

FROM redis:5.0.3-alpine

ENV PYTHON_VERSION=2.7.12-r1
ENV PY_PIP_VERSION=8.1.2-r0
ENV SUPERVISOR_VERSION=3.3.1

RUN apk update && apk add -u python py-pip
RUN pip install supervisor==$SUPERVISOR_VERSION

COPY --from=builder /target/pilot /usr/local/bin/autopilot-redis
RUN chmod +x /usr/local/bin/autopilot-redis
COPY supervisord.conf /etc/supervisord.conf

ENTRYPOINT ["supervisord"]
