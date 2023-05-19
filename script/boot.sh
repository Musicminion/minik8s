#!/bin/bash

SCRIPTS_ROOT="$(cd "$(dirname "$0")" && pwd)"

cd $SCRIPTS_ROOT/../

# 定义启动的程序列表，每个元素对应一个 main.go 文件和日志文件路径
programs=(
    "./pkg/apiserver/main/main.go:./log/apiserver.log"
    "./pkg/kubelet/main/main.go:./log/kubelet.log"
    "./pkg/scheduler/main/main.go:./log/scheduler.log"
    "./pkg/kubeproxy/main/main.go:./log/kubeproxy.log"
)

# 初始化测试环境
# 删除 etcd 中所有内容
. "$SCRIPTS_ROOT/etcd_clear.sh" /


# 删除除了weave之外的所有容器
. "$SCRIPTS_ROOT/container_clear.sh" /

# 清空iptables
echo "清空iptables"
iptables -t nat -F
iptables -t nat -X

# 重启weave
echo "重启weave"
weave stop
weave launch
weave expose

# 重启docker
echo "重启docker"
systemctl restart docker


# 循环启动程序
for program in "${programs[@]}"; do
    # 获取程序和日志文件路径
    IFS=':' read -ra ADDR <<< "$program"
    program_file="${ADDR[0]}"
    log_file="${ADDR[1]}"
    echo "启动程序：$program_file , 日志文件：$log_file"
    
    # 创建日志文件
    touch "$log_file"
    
    # 启动程序，并将标准输出和标准错误输出重定向到日志文件中
    sudo go run "$program_file" &> "$log_file" &
    # 如果是apiserver，需要sleep一段时间确保启动成功
    if [[ $program_file == *"apiserver"* ]]; then
        sleep 3
    fi
done

# 等待所有程序运行结束
wait