FROM golang:alpine AS build-env

COPY . /go/src/github.com/sapk/docker-volume-rclone
WORKDIR /go/src/github.com/sapk/docker-volume-rclone

RUN apk add --no-cache make git
RUN make clean build
RUN go get -u -v github.com/ncw/rclone

FROM alpine
LABEL maintainer="Antoine GIRARD <antoine.girard@sapk.fr>"

RUN apk add --no-cache ca-certificates bash fuse \
 && mkdir -p /var/lib/docker-volumes/rclone /etc/docker-volumes/rclone
COPY --from=build-env /go/src/github.com/sapk/docker-volume-rclone/docker-volume-rclone /usr/bin/docker-volume-rclone
COPY --from=build-env /go/bin/rclone /usr/bin/rclone

RUN /usr/bin/docker-volume-rclone version

ENTRYPOINT [ "/usr/bin/docker-volume-rclone" ]
CMD [ "daemon" ]
