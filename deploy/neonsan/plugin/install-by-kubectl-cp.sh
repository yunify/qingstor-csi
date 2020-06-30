#!/bin/bash

status=$(systemctl is-active neonsan-plugin)
if [[ ${status} = "active" ]]; then
  echo "neonsan-plugin is active, stop it "
  systemctl stop neonsan-plugin
fi

podName=$(kubectl -n kube-system get pod -l app=csi-neonsan,role=controller --field-selector=status.phase=Running -o name |  head -1 | awk -F'/' '{print $2}')
if test -z ${podName}; then
  echo "no controller pod, failed to start, please check"
  exit 1
fi

kubectl -n kube-system -c csi-neonsan cp ${podName}:neonsan-csi-driver /usr/bin/neonsan-plugin
chmod +x /usr/bin/neonsan-plugin

cat > /etc/systemd/system/neonsan-plugin.service  <<EOF
[Unit]
Description=NeonSAN CSI Plugin

[Service]
Type=simple
ExecStart=/usr/bin/neonsan-plugin  --endpoint=unix:///var/lib/kubelet/plugins/neonsan.csi.qingstor.com/csi.sock

Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable neonsan-plugin
systemctl start neonsan-plugin

status=$(systemctl is-active neonsan-plugin)
if [[ ${status} = "active" ]]; then
  echo "neonsan-plugin start successfully"
else
  echo "neonsan-plugin failed to start, please check"
fi
