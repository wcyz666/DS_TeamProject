#!/bin/bash
# argument 1 is the type you want to start up with
# argument 2 is the whether you want to clear dns
# Usage:
#	./run node
#	./run supernode clear
#	./run supernode
git pull
go build main.go || exit;

if [ "$2" != "" ]; then
	./main -clearDNS
fi

if [ "$1" == "supernode" ]; then
	./main -class=supernode
else
	./main -class=node
fi

