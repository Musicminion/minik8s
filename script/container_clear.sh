#!/bin/bash

echo "删除所有容器...(除了weave组件)"
# 获取当前机器上的所有容器ID
container_ids=$(docker ps -a -q)

# 遍历所有容器ID，删除名称不包含"weave"的容器
for id in $container_ids
do
    name=$(docker inspect --format '{{ .Name }}' $id)
    if [[ $name != *"weave"* ]]; then
        echo "删除容器 $name ..."
        docker rm -f $id
    else
        echo "跳过容器 $name ..."
    fi
done