FROM docker.io/library/golang:1.17.6 as builder
WORKDIR /app
ADD go.mod go.mod
ADD go.sum go.sum
RUN go mod download

ADD cmd cmd
RUN go build ./cmd/limbo

FROM docker.io/library/ubuntu:21.10
ENV DEBIAN_FRONTEND=noninteractive

RUN groupadd -g 999 limbo && \
    useradd -r -u 999 -g limbo limbo && \
    mkdir -p /home/limbo

COPY busybox.tar.gz /home/limbo
RUN chown -R limbo:0 /home/limbo && \
    chmod -R g=u /home/limbo

COPY --from=builder /app/limbo /usr/local/bin/limbo
ENV USER=limbo

USER 999
WORKDIR /home/limbo

ENTRYPOINT ["limbo"]
