#!/bin/bash

# 定义要杀死的进程列表
programs=(
    "apiserver"
    "kubelet"
    "scheduler"
    "kubeproxy"
    "controller"
    "serveless"
)

# 遍历进程列表，逐个杀死进程
for program in "${programs[@]}"
do
    # 提取进程名
    program_name=$(echo "$program" | awk -F '/' '{print $NF}')
  
    # 查找进程ID并杀死进程
    pid=$(ps aux | grep "$program_name" | grep -v grep | awk '{print $2}')
    if [ -n "$pid" ]; then
        echo "Killing process $pid for program $program_name..."
        kill -9 $pid
    else
        echo "Process for program $program_name not found"
    fi
done