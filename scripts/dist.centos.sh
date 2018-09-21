#!/bin/bash
set -e

# this script is used in a docker container, don't run it directly.

yum update -y && yum -y install zip wget tar

cd /app
mkdir -p centos && cd centos \
  && wget https://github.com/CyberMiles/libeni/releases/download/v1.3.2/libeni-1.3.2_centos-7.tgz -O libeni.tgz \
  && tar zxvf *.tgz --strip-components 1 && cp /app/travis .

zip "/scripts/travis_${BUILD_TAG}_centos-7.zip" travis lib/*
