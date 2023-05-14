#!/bin/bash
sudo wget -O /usr/local/bin/weave https://raw.githubusercontent.com/zettio/weave/master/weave 

# 记录下来当前的用户名
curUSER=$(whoami)
sudo su
# create systemd service file for weave
cat <<EOF > /etc/systemd/system/weave.service
[Unit]
Description=Weave Network
After=network.target

[Service]
EnvironmentFile=/etc/weave.env
ExecStart=/usr/local/bin/weave launch

RestartSec=10
StartLimitInterval=0

[Install]
WantedBy=multi-user.target
EOF

# create environment file for weave
cat <<EOF > /etc/weave.env
HOSTNAME=$(hostname -s)
EOF

# enable and start the weave service
sudo systemctl enable weave.service
sudo systemctl start weave.service

sudo su $curUSER