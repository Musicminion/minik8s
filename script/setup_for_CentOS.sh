#!/bin/bash


# 超级无敌巨无霸一键安装脚本

# 1s搞定环境安装 By zzq！

# 安装的东西有awk、docker、etcd、rabbitmq消息队列

# Check if awk is installed

if ! command -v awk &> /dev/null; then
    # Install awk
    echo "awk is not installed. Installing now..."
    sudo yum install -y awk
else
    echo "awk is already installed."
fi


# 检查Go是否已安装

if ! command -v go &> /dev/null
then
    echo "Go尚未安装。正在安装Go 1.20.3..."


    # 下载Go 1.20.3版本的二进制文件(使用aliyun镜像源)
    wget -c --tries=0 --timeout=300 --waitretry=5 --read-timeout=20 -O /tmp/go.tar.gz https://mirrors.aliyun.com/golang/go1.20.3.linux-amd64.tar.gz


    # 解压缩二进制文件到/usr/local目录
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz

    # 将Go二进制文件路径添加到PATH环境变量
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh

    # 使所有用户都能够访问PATH环境变量中的Go路径
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh > /dev/null
    sudo chmod +x /etc/profile.d/go.sh

    # 加载新的环境变量
    source /etc/profile.d/go.sh

    # 验证Go安装
    go version

else
    echo "Go已经安装。跳过安装步骤。"
    go version
fi


# 检查etcd是否已安装

if command -v etcd &> /dev/null
then
    echo "etcd已安装"
else
    # 如果etcd没有安装，则安装它
    echo "etcd未安装, 开始安装..."
    wget -q https://github.com/etcd-io/etcd/releases/download/v3.5.0/etcd-v3.5.0-linux-amd64.tar.gz >> /dev/null
    tar -xvf etcd-v3.5.0-linux-amd64.tar.gz
    cd etcd-v3.5.0-linux-amd64
    sudo mv etcd /usr/local/bin/
    sudo mv etcdctl /usr/local/bin/
    echo "etcd安装完成"


# 创建systemd启动脚本
    sudo tee /usr/lib/systemd/system/etcd.service > /dev/null <<EOF
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


检查RabbitMQ是否已经安装

if command -v rabbitmq-server &> /dev/null
then
    echo "RabbitMQ已安装"
else
    # 如果RabbitMQ没有安装，则安装它
    echo "RabbitMQ未安装，开始安装..."
    sudo yum install -y rabbitmq-server
    echo "RabbitMQ安装完成"


    # 设置RabbitMQ开机自动启动
    sudo systemctl enable rabbitmq-server

fi


# 启动RabbitMQ服务

sudo systemctl start rabbitmq-server

echo "RabbitMQ已安装并设置开机启动"


# 检查Docker是否已经安装

if [ ! -x "$(command -v docker)" ]; then


#安装必要的软件包以允许yum通过HTTPS使用存储库

  sudo yum install -y yum-utils device-mapper-persistent-data lvm2


# 设置Docker的存储库

  sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo


# 更新软件包列表

  sudo yum update -y


# 安装最新版本的Docker CE

  sudo yum install -y docker-ce docker-ce-cli containerd.io


# 将主机上的所有用户添加到Docker用户组

  sudo groupadd docker
  sudo gpasswd -a "$USER" docker
  sudo getent passwd | while IFS=: read -r name _ uid gid _ home shell; do
    [ $uid -ge 1000 ] && sudo gpasswd -a "$name" docker
  done


  echo "Docker安装完成并将主机上的所有用户添加到Docker用户组！注意！你需要手动启动机器"
else
  echo "Docker已经安装。"
fi