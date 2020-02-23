ARG RCLONE_VER=1.51

FROM golang:alpine AS build-env

COPY . /docker-volume-rclone
WORKDIR /docker-volume-rclone

RUN apk add --no-cache make git
RUN make clean build

FROM rclone/rclone:$RCLONE_VER
LABEL maintainer="Antoine GIRARD <antoine.girard@sapk.fr>"

RUN apk add --no-cache bash \
 && mkdir -p /var/lib/docker-volumes/rclone /etc/docker-volumes/rclone /var/cache/rclone \
 && ln -s /usr/local/bin/rclone /usr/bin/rclone
COPY --from=build-env /docker-volume-rclone/docker-volume-rclone /usr/local/bin/docker-volume-rclone

RUN /usr/local/bin/docker-volume-rclone version

ENTRYPOINT [ "/usr/local/bin/docker-volume-rclone" ]
CMD [ "daemon" ]
