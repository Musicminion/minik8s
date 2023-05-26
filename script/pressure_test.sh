#!/bin/bash

url="192.168.92.32:88"
num_requests=1000

for ((i=1; i<=$num_requests; i++))
do
    curl -s -o /dev/null $url &
done

wait