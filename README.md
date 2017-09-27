# rktup

## Deployment

```
sudo cp /path/to/rktup /usr/local/bin

sudo setcap cap_net_bind_service=+ep /usr/local/bin/rktup

cat <<EOF | sudo tee /etc/systemd/system/rktup.service
[Unit]
Description=rktup
Documentation=https://rktup.org

[Service]
User=nobody
Group=nogroup

ExecStart=/usr/local/bin/rktup --addr 127.0.0.1:33333 --hostname rktup.org --githubToken SECRET

PrivateTmp=true
PrivateDevices=true
ProtectHome=true
ProtectSystem=full

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable rktup
systemctl start rktup
```
