#!/usr/bin/env bash
set -e

BASEDIR=$(dirname $0)
cd ./${BASEDIR}
BASEDIR=$(pwd)

# setup cluster
mkdir -p ~/volumes
git clone https://github.com/CyberMiles/testnet.git ~/volumes/testnet

cd ~/volumes/testnet/travis/scripts
yes "" | sudo ./cluster.sh test 6 4
docker-compose up -d all
sleep 5
curl http://node-1:46657/status

# web3-cmt
git clone https://github.com/CyberMiles/web3-cmt.js ~/web3-cmt.js
cd ~/web3-cmt.js
git checkout dev
yarn install
yarn link

# integration test
cd $BASEDIR
yarn install
yarn link "web3-cmt"
yarn test
