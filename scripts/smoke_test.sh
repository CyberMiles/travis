#!/usr/bin/env bash
set -e

mkdir -p ~/volumes
git clone https://github.com/CyberMiles/testnet.git ~/volumes/testnet

cd ~/volumes/testnet/travis/scripts
yes "" | sudo ./cluster.sh test 6 4
docker-compose up -d all
sleep 5
curl http://node-1:46657/status

git clone https://github.com/CyberMiles/web3-cmt.js ~/web3-cmt.js
cd ~/web3-cmt.js
yarn install
yarn smoke
