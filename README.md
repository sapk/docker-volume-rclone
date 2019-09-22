# docker-volume-rclone [![License](https://img.shields.io/badge/license-MIT-red.svg)](https://github.com/sapk/docker-volume-rclone/blob/master/LICENSE) ![Project Status](http://img.shields.io/badge/status-beta-orange.svg)
[![GitHub release](https://img.shields.io/github/release/sapk/docker-volume-rclone.svg)](https://github.com/sapk/docker-volume-rclone/releases) [![Go Report Card](https://goreportcard.com/badge/github.com/sapk/docker-volume-rclone)](https://goreportcard.com/report/github.com/sapk/docker-volume-rclone)
[![codecov](https://codecov.io/gh/sapk/docker-volume-rclone/branch/master/graph/badge.svg)](https://codecov.io/gh/sapk/docker-volume-rclone)
 master : [![Travis master](https://api.travis-ci.org/sapk/docker-volume-rclone.svg?branch=master)](https://travis-ci.org/sapk/docker-volume-rclone) develop : [![Travis develop](https://api.travis-ci.org/sapk/docker-volume-rclone.svg?branch=develop)](https://travis-ci.org/sapk/docker-volume-rclone)


Use Rclone as a backend for docker volume. This permit to easely mount a lot of cloud provider (https://rclone.org/overview/).

Status : **BETA (work and in use but still need improvements)**

Use Rclone cli in the plugin container so it depend on fuse on the host.

## Docker plugin (Easy method) [![Docker Pulls](https://img.shields.io/docker/pulls/sapk/plugin-rclone.svg)](https://hub.docker.com/r/sapk/plugin-rclone) [![ImageLayers Size](https://img.shields.io/imagelayers/image-size/sapk/plugin-rclone/latest.svg)](https://hub.docker.com/r/sapk/plugin-rclone)
```
docker plugin install sapk/plugin-rclone
docker volume create --driver sapk/plugin-rclone --opt config="$(base64 ~/.config/rclone/rclone.conf)" --opt remote=some-remote:bucket/path --name test
docker run -v test:/mnt --rm -ti ubuntu
```

## Build
```
make
```

## Start daemon
```
./docker-volume-rclone daemon
OR in a docker container
docker run -d --device=/dev/fuse:/dev/fuse --cap-add=SYS_ADMIN --cap-add=MKNOD  -v /run/docker/plugins:/run/docker/plugins -v /var/lib/docker-volumes/rclone:/var/lib/docker-volumes/rclone:shared sapk/docker-volume-rclone
```

For more advance params : ```./docker-volume-rclone --help OR ./docker-volume-rclone daemon --help```
```
Run listening volume drive deamon to listen for mount request

Usage:
  docker-volume-rclone daemon [flags]

Global Flags:
  -b, --basedir string   Mounted volume base directory (default "/var/lib/docker-volumes/rclone")
  -v, --verbose          Turns on verbose logging
```

## Create and Mount volume
```
docker volume create --driver rclone --opt config="$(base64 ~/.config/rclone/rclone.conf)" --opt remote=some-remote:bucket/path --name test
docker run -v test:/mnt --rm -ti ubuntu
```

## Docker-compose
First put your rclone config in a env variable:
```
export RCLONE_CONF_BASE64=$(base64 ~/.config/rclone/rclone.conf)
```
And setup you docker-compose.yml file like that
```
volumes:
  some_vol:
    driver: sapk/plugin-rclone
    driver_opts:
      config: "${RCLONE_CONF_BASE64}"
      args: "--read-only --fast-list"
      remote: "some-remote:bucket/path"
```
You can also hard-code your config in the docker-compose file in place of the env variable.

## Inspired from :
 - https://github.com/ContainX/docker-volume-netshare/
 - https://github.com/vieux/docker-volume-sshfs/
 - https://github.com/sapk/docker-volume-gvfs
 - https://github.com/calavera/docker-volume-glusterfs
 - https://github.com/codedellemc/rexray

## How to debug docker managed plugin :
```
#Restart plugin in debug mode
docker plugin disable sapk/plugin-rclone
docker plugin set sapk/plugin-rclone DEBUG=1
docker plugin enable sapk/plugin-rclone

#Get files under /var/log of plugin
docker-runc --root /var/run/docker/plugins/runtime-root/plugins.moby list
docker-runc --root /var/run/docker/plugins/runtime-root/plugins.moby exec -t $CONTAINER_ID cat /var/log/rclone.log
docker-runc --root /var/run/docker/plugins/runtime-root/plugins.moby exec -t $CONTAINER_ID cat /var/log/docker-volume-rclone.log
```
