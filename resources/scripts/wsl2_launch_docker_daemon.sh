#! /bin/bash

DOCKER_DISTRO=$(cat /etc/*-release | grep DISTRIB_ID | sed 's/DISTRIB_ID=//g')
DOCKER_LOG_DIR=$HOME/docker_logs
mkdir -pm o=,ug=rwx "DOCKER_LOG_IDR"
/mnt/c/Windows/System32/wsl.exe -d $DOCKER_DISTRO sh -c "nohup sudo -b dockerd < /dev/null > $DOCKER_LOG_DIR/dockerd.log 2&1"