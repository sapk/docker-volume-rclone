ARG RCLONE_VER=1.50

FROM golang:alpine AS build-env

COPY . /docker-volume-rclone
WORKDIR /docker-volume-rclone

RUN apk add --no-cache make git
RUN make clean build

FROM rclone/rclone:$RCLONE_VER
LABEL maintainer="Antoine GIRARD <antoine.girard@sapk.fr>"

RUN apk add --no-cache bash \
 && mkdir -p /var/lib/docker-volumes/rclone /etc/docker-volumes/rclone /var/cache/rclone \
 && ln -s /root/rclone /usr/bin/rclone
COPY --from=build-env /docker-volume-rclone/docker-volume-rclone /usr/bin/docker-volume-rclone

RUN /usr/bin/docker-volume-rclone version

ENTRYPOINT [ "/usr/bin/docker-volume-rclone" ]
CMD [ "daemon" ]
