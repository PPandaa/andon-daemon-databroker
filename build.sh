#!/bin/bash
CONTAINER="ifps-andon-daemon-databroker"
DOCKER_REPO="iiicondor/$CONTAINER"
VERSION="dev"

docker build -t $DOCKER_REPO:$VERSION .
docker push $DOCKER_REPO:$VERSION
docker tag $DOCKER_REPO:$VERSION $DOCKER_REPO:demo
docker push $DOCKER_REPO:demo

docker rmi -f $(docker images | grep $CONTAINER | awk '{print $3}')