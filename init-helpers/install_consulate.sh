function install_consulate {

    set -ex

    readonly CONSULATE_VERSION="0.0.7"
    readonly CONSULATE_BIN_DIR="/usr/local/bin"
    readonly CONSULATE_USER=consulate
   
    # Create consulate user 
    id $CONSULATE_USER >& /dev/null || useradd --system $CONSULATE_USER

    # Download binary and add to PATH
    curl -L -o /tmp/consulate.tar.gz https://github.com/kadaan/consulate/releases/download/v${CONSULATE_VERSION}/consulate_linux_amd64.tar.gz
  (cd $CONSULATE_BIN_DIR ; tar -xf /tmp/consulate.tar.gz && rm /tmp/consulate.tar.gz)

    chmod +x $CONSULATE_BIN_DIR/consulate

    # Create and enable systemD target
    
  cat <<EOF > /etc/systemd/system/consulate.service
[Unit]
Description="Consulate - a middleware for consul health checks"
Documentation=https://github.com/kadaan/consulate
Requires=network-online.target
After=network-online.target
[Service]
Type=simple
User=$CONSULATE_USER
Group=$CONSULATE_USER
ExecStart=$CONSULATE_BIN_DIR/consulate server 
KillMode=process
Restart=on-failure
TimeoutSec=300s
LimitNOFILE=65536
EOF

  # SystemD reload and restart

  systemctl daemon-reload
  if systemctl is-enabled --quiet consulate; then
    systemctl reenable consulate
  else
    systemctl enable consulate
  fi

  if systemctl is-active --quiet consulate; then
    systemctl restart consulate
  else
    systemctl start consulate
  fi  

}
