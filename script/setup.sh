#!/bin/bash

# 超级无敌巨无霸一键安装脚本
# 1s搞定环境安装 By zzq！

# 安装的东西有awk、go、docker、etcd、rabbitmq消息队列、redis、weave网络插件
# 同时会对整个环境进行清理，删除etcd所有内容，并删除除了weave之外的所有container

# Check if awk is installed
if ! command -v awk &> /dev/null; then
    # Install awk
    echo "awk is not installed. Installing now..."
    sudo apt-get update
    sudo apt-get install -y awk
else
    echo "awk is already installed."
fi


# 检查Go是否已安装
if ! command -v go &> /dev/null
then
    echo "Go尚未安装。正在安装Go 1.20.3..."

    # 下载Go 1.20.3版本的二进制文件
    wget --tries=0 --timeout=300 --waitretry=5 --read-timeout=20 https://mirrors.aliyun.com/golang/go1.20.3.linux-amd64.tar.gz  -O /tmp/go.tar.gz

    # 解压缩二进制文件到/usr/local目录
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz

    # 将Go二进制文件路径添加到PATH环境变量
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh

    # 使所有用户都能够访问PATH环境变量中的Go路径
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh > /dev/null
    sudo chmod 777 /etc/profile.d/go.sh

    # 加载新的环境变量
    source /etc/profile.d/go.sh

    # 把Defaults secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin:/usr/local/go/bin/"添加到/etc/sudoers文件中
    sudo sed -i 's/secure_path="/secure_path="\/usr\/local\/go\/bin:/' /etc/sudoers
    
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
  sudo gpasswd -a "$USER" docker
  sudo getent passwd | while IFS=: read -r name _ uid gid _ home shell; do
    [ $uid -ge 1000 ] && sudo gpasswd -a "$name" docker
  done

  echo "Docker安装完成并将主机上的所有用户添加到Docker用户组！注意！你需要手动启动机器"
else
  echo "Docker已经安装。"
fi

# 安装Redis
if command -v redis-server &> /dev/null
then
    echo "Redis已安装,尝试启动..."
    sudo systemctl start redis-server
else
    # 如果Redis没有安装，则安装它
    echo "Redis未安装, 开始安装..."
    sudo apt-get update
    sudo apt install -y lsb-release

    curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg

    echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list

    sudo apt-get update
    sudo apt-get install -y redis
    echo "Redis安装完成"

    # 设置Redis开机自动启动
    sudo systemctl enable redis-server
    # 启动Redis服务
    sudo systemctl start redis-server
fi

# 安装Weave网络插件
if command -v weave &> /dev/null
then
    echo "Weave已安装"
else
    # 如果Weave没有安装，则安装它
    echo "Weave未安装，开始安装..."
    # 下载Weave二进制文件
    sudo wget -O /usr/local/bin/weave https://raw.githubusercontent.com/zettio/weave/master/weave && sudo chmod +x /usr/local/bin/weave

    # 启动Weave网络
    sudo weave launch
    echo "Weave安装完成"
fi

# 获取脚本所在目录
SCRIPTS_ROOT="$(cd "$(dirname "$0")" && pwd)"

### 以下内容用于格式化服务器的部分数据
# 删除 etcd 中所有内容
. "$SCRIPTS_ROOT/etcd_clear.sh" /


# 删除除了weave之外的所有容器
. "$SCRIPTS_ROOT/container_clear.sh" /

# 删除相关进程
. "$SCRIPTS_ROOT/process_clear.sh" /

# 设置项目的环境变量
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export MINIK8S_PATH="$PROJECT_ROOT"
echo "设置环境变量: MINIK8S_PATH=$MINIK8S_PATH"
