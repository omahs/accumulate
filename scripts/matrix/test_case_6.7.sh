#!/bin/bash
#
# test case 6.7
#
# xfer funds between lite accounts, one that has no funds to transfer
# ids, amount and server IP:Port needed
#
# set cli command and see if it exists
#
export cli=../../cmd/cli/cli

if [ ! -f $cli ]; then
        echo "cli command not found in ../../cmd/cli, attempting to build"
        ./build_cli.sh
        if [ ! -f $cli ]; then
           echo "cli command failed to build"
           exit 1
        fi
fi

# create some IDs and don't faucet either of them

id1=`cli_create_id.sh $1`
if [ $? -ne 0 ]; then
	echo "cli create id failed"
	exit 1
fi
id2=`cli_create_id.sh $1`
if [ $? -ne 0 ]; then
	echo "cli create id failed"
	exit 1
fi

#call our xfer script

./cli_xfer_tokens.sh $id1 $id2 $3 $4

if [ $? -ne 0 ]; then
	echo "cli xfer tokens failed"
	exit 1
fi
exit 0
