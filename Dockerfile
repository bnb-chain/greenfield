FROM golang:1.21-alpine as builder

RUN apk add --no-cache make git bash build-base linux-headers libc-dev

ADD . /greenfield

ENV CGO_ENABLED=1
ENV GO111MODULE=on

# For Private REPO
ARG GH_TOKEN=""
RUN go env -w GOPRIVATE="github.com/bnb-chain/*"
RUN git config --global url."https://${GH_TOKEN}@github.com".insteadOf "https://github.com"

RUN cd /greenfield && make build

# Pull greenfield into a second stage deploy alpine container
FROM alpine:3.17

ARG USER=greenfield
ARG USER_UID=1000
ARG USER_GID=1000

ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"
ENV CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"

ENV PACKAGES ca-certificates libstdc++
ENV WORKDIR=/app

RUN apk add --no-cache $PACKAGES \
  && rm -rf /var/cache/apk/* \
  && addgroup -g ${USER_GID} ${USER} \
  && adduser -u ${USER_UID} -G ${USER} --shell /sbin/nologin --no-create-home -D ${USER} \
  && addgroup ${USER} tty \
  && sed -i -e "s/bin\/sh/bin\/bash/" /etc/passwd


WORKDIR ${WORKDIR}

COPY --from=builder /greenfield/build/bin/gnfd ${WORKDIR}/
RUN chown -R ${USER_UID}:${USER_GID} ${WORKDIR}
USER ${USER_UID}:${USER_GID}

EXPOSE 26656 26657 9090 1317 6060 4500

ENTRYPOINT ["/app/gnfd"]
