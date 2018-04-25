#!/bin/bash

INST_COUNT=$1
CLEAR=$2
CLS='--clear'
MV='--mv' #multi validators

BASE_DIR=~/.travis

TRPCPORT=46657
TP2PPORT=46656
ERPCPORT=8545

RSTEP=10

seeds="127.0.0.1:$TP2PPORT"

# init params
if [ -z $INST_COUNT ]; then
	INST_COUNT=1
	CLEAR='-'
else
	if [[ ! $INST_COUNT =~ ^[0-9]+$ ]]; then
		if [ $INST_COUNT == "$CLS" ]; then
			CLEAR="$CLS"
		fi
		if [ $INST_COUNT == "$MV" ]; then
			CLEAR="$MV"
		fi
		INST_COUNT=1
	else
		if [ -z $CLEAR ]; then
			CLEAR='-'
		fi
	fi
fi

# init seeds
# for now only 46656 one node included
if [ $INST_COUNT -gt 1 ]; then
	for i in `seq 2 $INST_COUNT`
	do
		tp2pport=$(($TP2PPORT + $i * $RSTEP))
		# seeds="$seeds,127.0.0.1:$tp2pport"
	done
fi

# change several ports to specific with step as 10
modifyConf()
{
	dir=$1
	seq=$2

	trpcport=$(($TRPCPORT + $seq * $RSTEP))
	tp2pport=$(($TP2PPORT + $seq * $RSTEP))
	erpcport=$(($ERPCPORT + $seq * $RSTEP))


	cd $dir
	cp $BASE_DIR/genesis.json .
	sed -i '' "s/$TRPCPORT/$trpcport/g" ./config.toml
	sed -i '' "s/$TP2PPORT/$tp2pport/g" ./config.toml
	sed -i '' "s/$ERPCPORT/$erpcport/g" ./config.toml
	sed -i '' "s/seeds = \"\"/seeds = \"$seeds\"/g" ./config.toml
}

# kill running travis first
rc=`ps aux | grep "[t]ravis node start" | wc -l`
if [ $rc -ne 0 ] ; then
	ps aux | grep "[t]ravis node start" | awk '{print $2}' | xargs kill
fi

while true
do
	c=`ps aux | grep "[t]ravis node start" | awk '{print $2}' | wc -l`
	if [ $c -ne 0 ]; then
		echo 'stopping old travis nodes'
		sleep 1
	else
		break
	fi
done

if [ $INST_COUNT -eq 0 ]; then
	exit
fi

cd

for i in `seq 1 $INST_COUNT`
do
	if [ $i -ne 1 ]; then
		# seq is empty string for the first instance
		seq=$(($i - 1))
	fi

	dir=$BASE_DIR$seq

	# make .travis* directory if not exist
	if [ ! -d "$dir" ]; then
		mkdir $dir
		newnode=1
	else
		newnode=0
	fi

	cd $dir

	if [ $CLEAR == "$CLS" ] || [ $newnode -eq 1 ] ; then
		rm -rf *
		travis node init --home .
		if [ $i -ne 1 ]; then
			modifyConf $dir $seq
		else
			sed -i '' "s/seeds = \"\"/seeds = \"$seeds\"/g" ./config.toml
		fi
	fi

	if [ $CLEAR == "$MV" ] ; then
		# ....
		cd ..
		rm -rf $dir
		cp -r ~/.o_travis/.travis$seq ~
		cd $dir
		# ....
	fi

	if [ $INST_COUNT -eq 1 ]; then
		travis node start --home .
	else
		travis node start --home . > travis.log 2>&1 &
	fi
done