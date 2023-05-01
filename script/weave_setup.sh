#!/bin/bash

# 定义主机列表
hosts=("192.168.118.132")

# 遍历主机列表进行安装和运行
for host in "${hosts[@]}"
do
    # 安装Weave
    ssh $host "sudo wget -O /usr/local/bin/weave https://raw.githubusercontent.com/zettio/weave/master/weave && sudo chmod +x /usr/local/bin/weave"
done