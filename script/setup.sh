#!/bin/bash

# 超级无敌巨无霸一键安装脚本
# 1s搞定环境安装 By zzq！

# 安装的东西有awk、docker、etcd、rabbitmq消息队列


# Check if awk is installed
if ! command -v awk &> /dev/null; then
    # Install awk
    echo "awk is not installed. Installing now..."
    sudo apt-get update
    sudo apt-get install -y awk
else
    echo "awk is already installed."
fi


# 检查是否已经安装Go
if ! command -v go &> /dev/null
then
    echo "Go Need install"
    cd
    wget -q https://go.dev/dl/go1.20.3.linux-amd64.tar.gz >> /dev/null
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

# Set the Go installation directory
GO_INSTALL_DIR=/usr/local/go

# Loop through all user accounts
for username in $(awk -F: '{print $1}' /etc/passwd); do
  # Get the home directory for the user
  homedir=$(eval echo ~$username)

  # Check if the .bashrc file exists for the user
  if [ -f "$homedir/.bashrc" ]; then
    # Add the GOPATH and PATH environment variables to the .bashrc file
    echo "export GOPATH=$homedir/go" >> "$homedir/.bashrc"
    echo "export PATH=\$PATH:\$GOPATH/bin:$GO_INSTALL_DIR/bin" >> "$homedir/.bashrc"

    # Load the new environment variables for the current session
    source "$homedir/.bashrc"
  fi
done


# 检查etcd是否已安装
if command -v etcd &> /dev/null
then
    echo "etcd已安装"
else
    # 如果etcd没有安装，则安装它
    echo "etcd未安装，开始安装..."
    wget -q https://github.com/etcd-io/etcd/releases/download/v3.5.0/etcd-v3.5.0-linux-amd64.tar.gz >> /dev/null
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

# 检查Docker是否已经安装
if [ ! -x "$(command -v docker)" ]; then
  # 更新软件包列表
  sudo apt-get update

  # 安装必要的软件包以允许apt通过HTTPS使用存储库
  sudo apt-get install -y \
      apt-transport-https \
      ca-certificates \
      curl \
      gnupg-agent \
      software-properties-common

  # 添加Docker的官方GPG密钥
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

  # 添加Docker的存储库
  sudo add-apt-repository \
     "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
     $(lsb_release -cs) \
     stable"

  # 更新软件包列表
  sudo apt-get update

  # 安装最新版本的Docker CE
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io

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