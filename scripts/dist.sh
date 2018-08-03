#!/bin/bash
set -e

# this script is used in a docker container, don't run it directly.

apt-get update && apt-get -y install zip wget tar

# ubuntu
cd /app
zip "/scripts/travis_${BUILD_TAG}_ubuntu-16.04.zip" travis lib/*

# centos
mkdir -p centos && cd centos \
  && wget https://github.com/CyberMiles/libeni/releases/download/v1.2.0/libeni-1.2.0_centos-7.tgz -O libeni.tgz \
  && tar zxvf *.tgz --strip-components 1 && cp /app/travis .

zip "/scripts/travis_${BUILD_TAG}_centos-7.zip" travis lib/*
