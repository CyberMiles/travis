# Travis
[![Build Status develop branch](https://travis-ci.org/CyberMiles/travis.svg?branch=develop)](https://travis-ci.org/CyberMiles/travis)

Please see the documentation for building and deploying Travis nodes here: https://travis.readthedocs.io/en/latest/getting-started.html

## Automated Script - CyberMiles testnet
Below is a bash script which will get Lity, libENI and a Travis node running on Ubuntu 16.04. Disclaimer: These instructions are for a brand-new disposable 16.04 Ubuntu test instance which has the sole purpose of running and testing the CyberMiles Lity compiler, CyberMiles libENI framework and the CyberMiles testnet called Travis.

```
#/bin/bash

#Disclaimer: These instructions are for a brand-new disposable 16.04 Ubuntu test instance which has the sole purpose of running and testing the CyberMiles Lity compiler, CyberMiles libENI framework and the CyberMiles testnet called Travis.

#To use this file copy this text into a file called install_cybermiles.sh in your fresh disposable Ubuntu 16.04 machine's home directory. Make the file executable by running "sudo chmod a+x ~/install_cybermiles.sh" and then finally execute the file by running "cd ~" and then "./install_cybermiles.sh"

#System preparation Ubuntu 16.04 LTS
cd ~
sudo apt-get -y update
sudo apt-get -y upgrade
sudo apt-get -y autoremove
sudo apt-get -y install gcc
sudo apt-get -y install git
sudo apt-get -y install make
sudo apt-get -y install curl
sudo apt-get -y install wget
sudo apt-get -y install cmake
sudo apt-get -y install bison
sudo apt-get -y install openssl
sudo apt-get -y install binutils
sudo apt-get -y install automake
sudo apt-get -y install libssl-dev
sudo apt-get -y install libboost-dev
sudo apt-get -y install libaudit-dev
sudo apt-get -y install libblkid-dev
sudo apt-get -y install e2fslibs-dev
sudo apt-get -y install build-essential
sudo apt-get -y install libboost-all-dev

#Installing Lity
cd ~
git clone https://github.com/CyberMiles/lity.git
cd lity
mkdir build
cd build
cmake ..
make
./lityc/lityc --help

#Installing SkyPat
cd ~
wget https://github.com/skymizer/SkyPat/archive/v3.1.1.tar.gz
tar -zxvf v3.1.1.tar.gz
cd SkyPat-3.1.1
./autogen.sh
./configure
make
sudo make install

#Installing Go
cd ~
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source $HOME/.gvm/scripts/gvm
gvm install go1.10.3 -B
gvm use go1.10.3 --default
echo 'export GOPATH=~/.gvm/pkgsets/go1.10.3/global' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/go/bin' >> ~/.bashrc
echo 'export PATH=$GOBIN:$PATH' >> ~/.bashrc
source ~/.bashrc

#Installing Travis
go get github.com/CyberMiles/travis

#PLEASE NOTE: The above will return an error such as "can't load package ... no Go file in ...",
#Please just ignore this error and continue on with the installation

cd $GOPATH/src/github.com/CyberMiles/travis
git checkout master

#Incorporate libENI
sudo rm -rf ~/.travis
wget -O ~/libeni.tgz https://github.com/CyberMiles/libeni/releases/download/v1.3.4/libeni-1.3.4_ubuntu-16.04.tgz
tar zxvf ~/libeni.tgz -C ~
mkdir -p ~/.travis/eni
cp -r ~/libeni-1.3.4/lib ~/.travis/eni/lib

#Continue installing Travis
cd ~
cd $GOPATH/src/github.com/CyberMiles/travis
make all

#Configuring Travis test network settings
cd ~
git clone https://github.com/CyberMiles/testnet.git
cp -r testnet/travis/init .travis

#Starting Travis test network node
cd ~
travis node init --env testnet
curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/config.toml > ~/.travis/config/config.toml
curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/genesis.json > ~/.travis/config/genesis.json


#Please ensure that the system paths are known, or else the travis command will not be found (you will get an error like this "The program 'travis' is currently not installed")

echo 'export GOPATH=~/.gvm/pkgsets/go1.10.3/global' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/go/bin' >> ~/.bashrc
echo 'export PATH=$GOBIN:$PATH' >> ~/.bashrc
source ~/.bashrc

travis node start --home=./.travis

```

You can now attach to the Travis node using the following command

```
travis attach http://localhost:8545
```
## Automated Script - CyberMiles PRIVATE network

Below is a bash script which will get Lity, libENI and a Travis node running on Ubuntu 16.04, as a private network. Running a private network is useful for when you want to perform rapid experimentation; when you do not want to wait to sync with a mainnet or testnet. 

Disclaimer: These instructions are for a brand-new disposable 16.04 Ubuntu test instance which has the sole purpose of running and testing the CyberMiles Lity compiler, CyberMiles libENI framework and the CyberMiles testnet called Travis. The following script will initialize a new private blockchain starting at block 0. 
```
#/bin/bash

#Disclaimer: These instructions are for a brand-new disposable 16.04 Ubuntu test instance which has the sole purpose of running and testing the CyberMiles Lity compiler, CyberMiles libENI framework and the CyberMiles testnet called Travis.

#To use this file copy this text into a file called install_cybermiles_private_network.sh in your fresh disposable Ubuntu 16.04 machine's home directory. Make the file executable by running "sudo chmod a+x ~/install_cybermiles_private_network.sh" and then finally execute the file by running "cd ~" and then "./install_cybermiles_private_network.sh"

#System preparation Ubuntu 16.04 LTS
cd ~
sudo apt-get -y update
sudo apt-get -y upgrade
sudo apt-get -y autoremove
sudo apt-get -y install gcc
sudo apt-get -y install git
sudo apt-get -y install make
sudo apt-get -y install curl
sudo apt-get -y install wget
sudo apt-get -y install cmake
sudo apt-get -y install bison
sudo apt-get -y install openssl
sudo apt-get -y install binutils
sudo apt-get -y install automake
sudo apt-get -y install libssl-dev
sudo apt-get -y install libboost-dev
sudo apt-get -y install libaudit-dev
sudo apt-get -y install libblkid-dev
sudo apt-get -y install e2fslibs-dev
sudo apt-get -y install build-essential
sudo apt-get -y install libboost-all-dev

#Installing Lity
cd ~
git clone https://github.com/CyberMiles/lity.git
cd lity
mkdir build
cd build
cmake ..
make
./lityc/lityc --help

#Installing SkyPat
cd ~
wget https://github.com/skymizer/SkyPat/archive/v3.1.1.tar.gz
tar -zxvf v3.1.1.tar.gz
cd SkyPat-3.1.1
./autogen.sh
./configure
make
sudo make install

#Installing Go
cd ~
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source $HOME/.gvm/scripts/gvm
gvm install go1.10.3 -B
gvm use go1.10.3 --default
echo 'export GOPATH=~/.gvm/pkgsets/go1.10.3/global' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/go/bin' >> ~/.bashrc
echo 'export PATH=$GOBIN:$PATH' >> ~/.bashrc
source ~/.bashrc

#Installing Travis
go get github.com/CyberMiles/travis

#PLEASE NOTE: The above will return an error such as "can't load package ... no Go file in ...",
#Please just ignore this error and continue on with the installation

cd $GOPATH/src/github.com/CyberMiles/travis
git checkout master

#Incorporate libENI
sudo rm -rf ~/.travis
wget -O ~/libeni.tgz https://github.com/CyberMiles/libeni/releases/download/v1.3.4/libeni-1.3.4_ubuntu-16.04.tgz
tar zxvf ~/libeni.tgz -C ~
mkdir -p ~/.travis/eni
cp -r ~/libeni-1.3.4/lib ~/.travis/eni/lib

#Continue installing Travis
cd ~
cd $GOPATH/src/github.com/CyberMiles/travis
make all

#Starting Travis test network node
cd ~
travis node init

#Please ensure that the system paths are known, or else the travis command will not be found (you will get an error like this "The program 'travis' is currently not installed")

echo 'export GOPATH=~/.gvm/pkgsets/go1.10.3/global' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/go/bin' >> ~/.bashrc
echo 'export PATH=$GOBIN:$PATH' >> ~/.bashrc
source ~/.bashrc

travis node start
```

Obviously with a private network there is no syncing required because the blockchain starts at 0 on your own private hardware. You can now attach to the Travis node using the following command.

```
travis attach http://localhost:8545
``` 

There is a preconfigured coinbase account in the private network. You can unlock it using the default password as shown below. Once unlocked you can send value from this evergrowing coinbase account to other new accounts and then get on with your testing and development.

```
personal.unlockAccount(cmt.coinbase, "1234")
```
