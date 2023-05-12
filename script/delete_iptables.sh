#!/bin/bash

# 检查是否为 root 用户

if [[ $EUID -ne 0 ]]; then
    echo "需要 root 权限执行此脚本"
    exit 1
fi


# 检查是否有参数

if [[ $# -eq 0 ]]; then
    echo "请输入您要搜索的关键字"
    exit 1
fi


# 搜索规则和 Chain

results="$(iptables -L -n -v --line-numbers | grep -i "$1" && echo -e "\n")"
results+="$(iptables -t nat -L -n -v --line-numbers | grep -i "$1" && echo -e "\n")"
if [[ -z $results ]]; then
    echo "找不到任何与 '$1' 相关的规则或 Chain"
    exit 1
else
    echo -e "$results"
fi

# 确认是否删除搜索结果

read -p "是否删除以上搜索结果？[y/n] " confirm
if [[ $confirm == "y" ]]; then
    # 删除规则和 Chain
    iptables-save | grep -v $1 | iptables-restore
    echo "已删除匹配 $1 的所有规则和 Chain"
else
    echo "已取消操作"
fi


exit 0