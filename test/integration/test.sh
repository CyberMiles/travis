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
sed -i 's/image:/privileged: true\n    image:/g' docker-compose.yml
docker-compose up -d all
sleep 3
curl http://localhost:26657/status

# web3-cmt
git clone https://github.com/CyberMiles/web3-cmt.js ~/web3-cmt.js
cd ~/web3-cmt.js
git checkout master
yarn install
yarn link

# integration test
cd $BASEDIR
yarn install
yarn link "web3-cmt"
yarn test

# cleanup
cd ~/volumes/testnet/travis/scripts
docker-compose down
