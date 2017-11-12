FROM golang:alpine AS build-env
COPY . /go/src/app
WORKDIR /go/src/app

RUN apk add --no-cache make \
 && make build

FROM golang:alpine AS build-rclone
RUN go get -u -v github.com/ncw/rclone

FROM alpine
MAINTAINER Antoine GIRARD <antoine.girard@sapk.fr>

RUN apk add --no-cache ca-certificates bash fuse \
 && mkdir -p /var/lib/docker-volumes/rclone /etc/docker-volumes/rclone

COPY --from=build-rclone /go/bin/rclone /usr/bin/rclone
COPY --from=build-env /go/src/app/docker-volume-rclone /usr/bin/docker-volume-rclone

RUN /usr/bin/docker-volume-rclone version

ENTRYPOINT [ "/usr/bin/docker-volume-rclone" ]
CMD [ "daemon" ]
