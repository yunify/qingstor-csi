#!/usr/bin/env bash

cat <<EOF > /etc/systemd/system/neonsan-plugin.service
[Unit]
Description=NeonSAN CSI Plugin

[Service]
Type=simple
ExecStart=/usr/bin/neonsan-plugin  \
           --endpoint=unix:///var/lib/kubelet/plugins/neonsan.csi.qingcloud.com/csi.sock

Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
EOF

systemctl enable neonsan-plugin.service
systemctl start neonsan-plugin
