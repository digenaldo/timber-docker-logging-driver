#!/usr/bin/env bash

docker build -t rootfsimage .
id=$(docker create rootfsimage true)
mkdir rootfs
docker export "$id" | tar -x -C rootfs

docker rm -vf "$id"
docker rmi rootfsimage

docker plugin create timber .
docker plugin enable timber
