#! /bin/bash
source ./config
if [ "$1" != "" ]
then
 export CS4032_FS_PORT=$1
fi
./directoryserver
