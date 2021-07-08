#!/bin/bash
CONTAINER="ifps-andon-daemon-databroker"
DOCKER_REPO="iiicondor/$CONTAINER"
VERSION="1.0.16"

docker build -t $DOCKER_REPO:$VERSION .
docker push $DOCKER_REPO:$VERSION
MESSAGE="delete IFPS_ANDON_UI_URL"
echo "[`date "+%Y-%m-%d %H:%M:%S"`] $VERSION => On-Premise {$MESSAGE}" >> ImageInfo.txt

# docker build -t $DOCKER_REPO:$VERSION .
# docker push $DOCKER_REPO:$VERSION
# docker tag $DOCKER_REPO:$VERSION $DOCKER_REPO:dev
# docker push $DOCKER_REPO:dev
# MESSAGE="annotation dashboardPrincipal()"
# echo "[`date "+%Y-%m-%d %H:%M:%S"`] $VERSION => dev {$MESSAGE}" >> ImageInfo.txt

# docker pull $DOCKER_REPO:$VERSION
# docker tag $DOCKER_REPO:$VERSION $DOCKER_REPO:demo
# docker push $DOCKER_REPO:demo
# echo "[`date "+%Y-%m-%d %H:%M:%S"`] $VERSION => demo" >> ImageInfo.txt

docker rmi -f $(docker images | grep $DOCKER_REPO | awk '{print $3}')
docker image prune -f