#! /bin/bash
source ./config
if [ "$1" != "" ]
then
 export CS4032_FS_PORT=$1
fi
if [ "$2" != "" ]
then
 export CS4032_FS_NODE=$2
fi
if [ "$3" != "" ]
then
 export CS4032_FS_DIRECTORY=$3
fi
./fileserver
