#!/bin/bash

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export MINIK8S_PATH="$PROJECT_ROOT"

SCRIPTS_ROOT="$(cd "$(dirname "$0")" && pwd)"

# 初始化测试环境
# 删除 etcd 中所有内容
. "$SCRIPTS_ROOT/worker_remake.sh" /


cd $PROJECT_ROOT

mkdir -p log

# 定义启动的程序列表，每个元素对应一个 main.go 文件和日志文件路径
programs=(
    "./pkg/kubelet/main/main.go:./log/kubelet.log"
    "./pkg/kubeproxy/main/main.go:./log/kubeproxy.log"
)

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
    # 如果是apiserver或者kubelet，需要sleep一段时间确保启动成功
    if [[ "$program_file" == *"apiserver"* ]] || [[ "$program_file" == *"kubelet"* ]]; then
        sleep 5
    fi
done

# 等待所有程序运行结束
wait