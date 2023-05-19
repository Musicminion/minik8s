#!/bin/bash

chainList=$(iptables -L -t nat | grep -vE "^.* (PREROUTING|INPUT|OUTPUT|POSTROUTING) .*"  | grep "Chain" | awk '{print $2}' | tac )
allChain=$(iptables -L -t nat    | grep "Chain" | awk '{print $2}' | tac )
for i in $allChain
do
    iptables -t nat -F $i
done

for i in $chainList
do
    iptables -t nat -X $i
done
