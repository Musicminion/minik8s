#!/bin/bash

SCRIPTS_ROOT="$(cd "$(dirname "$0")" && pwd)"

# 清空Redis
. "$SCRIPTS_ROOT/redis_clear.sh"

# 删除除了weave之外的所有容器
. "$SCRIPTS_ROOT/container_clear.sh" /

# 清空iptables
echo "清空iptables"
. "$SCRIPTS_ROOT/iptables_clear.sh" 

# 删除相关进程
. "$SCRIPTS_ROOT/process_clear.sh" /