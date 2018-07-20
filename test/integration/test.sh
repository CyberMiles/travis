#!/usr/bin/env bash
set -e

BASEDIR=$(dirname $0)
cd ./${BASEDIR}
BASEDIR=$(pwd)

# setup cluster
mkdir -p ~/volumes
git clone https://github.com/CyberMiles/testnet.git ~/volumes/testnet

cd ~/volumes/testnet/travis/scripts
git checkout master
yes "" | sudo ./cluster.sh test 6 4
docker-compose up -d all
sleep 3
curl http://localhost:26657/status

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


# single node test
cd ~/volumes/testnet/travis/scripts
docker-compose down

IMG=ywonline/travis
docker run --rm -v ~/volumes/local:/travis $IMG node init --home=/travis
docker run --rm -v ~/volumes/local:/travis -d -p 26657:26657 -p 8545:8545 $IMG node start --home=/travis
sleep 3
curl http://localhost:26657/status

cd $BASEDIR
yarn test
