#!/bin/bash

# 检查是否已经安装Go
if ! command -v go &> /dev/null
then
    echo "Go Need install"
    cd
    wget https://go.dev/dl/go1.20.3.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.20.3.linux-amd64.tar.gz
    sudo su
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    exit
    source /etc/profile
    rm go1.20.3.linux-amd64.tar.gz

    # 在文章的末尾追加,保证sudo用户可以搞
    echo 'Defaults secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin:/usr/local/go/bin/"' >> /etc/sudoers
else
    echo "Go Already install "
fi

# 检查etcd是否已安装
if command -v etcd &> /dev/null
then
    echo "etcd已安装"
else
    # 如果etcd没有安装，则安装它
    echo "etcd未安装，开始安装..."
    wget https://github.com/etcd-io/etcd/releases/download/v3.5.0/etcd-v3.5.0-linux-amd64.tar.gz
    tar -xvf etcd-v3.5.0-linux-amd64.tar.gz
    cd etcd-v3.5.0-linux-amd64
    sudo mv etcd /usr/local/bin/
    sudo mv etcdctl /usr/local/bin/
    echo "etcd安装完成"

    # 创建systemd启动脚本
    sudo tee /etc/systemd/system/etcd.service > /dev/null <<EOF
    [Unit]
    Description=etcd
    Documentation=https://github.com/etcd-io/etcd
    After=network-online.target

    [Service]
    User=root
    Type=notify
    ExecStart=/usr/local/bin/etcd

    [Install]
    WantedBy=multi-user.target
EOF

    # 重新加载systemd配置
    sudo systemctl daemon-reload

    # 设置开机自动启动etcd
    sudo systemctl enable etcd
fi

# 启动etcd服务
sudo systemctl start etcd
echo "etcd已安装并设置开机启动"


# 检查RabbitMQ是否已经安装
if command -v rabbitmq-server &> /dev/null
then
    echo "RabbitMQ已安装"
else
    # 如果RabbitMQ没有安装，则安装它
    echo "RabbitMQ未安装，开始安装..."
    sudo apt-get update
    sudo apt-get install rabbitmq-server -y
    echo "RabbitMQ安装完成"

    # 设置RabbitMQ开机自动启动
    sudo systemctl enable rabbitmq-server
fi

# 启动RabbitMQ服务
sudo systemctl start rabbitmq-server

echo "RabbitMQ已安装并设置开机启动"

