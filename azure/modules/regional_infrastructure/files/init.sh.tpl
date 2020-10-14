#!/usr/bin/env bash

### Default trap for all errors in this script
default_trap () 
{
  echo "Catching error on line $LINENO. Shutting down instance";
  shutdown -P +1
  curl -i -XPOST 'http://localhost:12500/telegraf' --data-binary "health value=100000"
  /usr/local/bin/consul leave; 
}

# Set verbose mode and quit on error flags
set -x -eE

# Get hostname
INSTANCE_ID=$(hostname)

# Install custom repos
curl -o /etc/yum.repos.d/azure-cli.repo -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/azure/azure-cli.repo 
curl -o /etc/yum.repos.d/influxdb.repo -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/azure/influxdb.repo 

### Start of main script
# Install unzip, docker, jq, awscli
# /usr/bin/yum update -y
# /usr/bin/yum check-update -y
/usr/bin/rpm --import https://packages.microsoft.com/keys/microsoft.asc
set +eE
/usr/bin/rpm -Uvh https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm
set -eE
/usr/bin/yum install telegraf unzip docker jq azure-cli git nc -y
/usr/bin/az login --identity --allow-no-subscriptions

# Make file system on a disk without force. Allow error. Error will be thrown if the disk already has filesystem. This will prevent data erasing.
set +eE
/usr/sbin/mkfs.xfs /dev/sdc
/usr/bin/mkdir -p /data
if grep -qs '/data ' /proc/mounts; then
    echo "Disk already mounted, skipping..."
else
    mount /dev/sdc /data
fi
chown polkadot:polkadot /data -R
set -eE

trap default_trap ERR EXIT
# Mound disk and add automount entry in /etc/fstab
/usr/bin/grep -qxF "/dev/sdc /data xfs defaults 0 0" /etc/fstab || /usr/bin/echo "/dev/sdc /data xfs defaults 0 0" >> /etc/fstab

# Run docker with regular polkadot container inside of it
/usr/bin/systemctl enable docker
/usr/bin/systemctl start docker

# Fetch validator instance parameters from Secret Manager
exit_code=1
until [ $exit_code -eq 0 ]; do

  trap - ERR
  set +eE
  az keyvault secret show --name "polkadot-${prefix}-name" --vault-name "${key_vault_name}"
  exit_code=$?
  set -eE
  trap default_trap ERR
  echo "Can not access secret at vault. Sleeping for 30s..."
  sleep 30

done

NAME=$(az keyvault secret show --name "polkadot-${prefix}-name" --vault-name "${key_vault_name}" | jq .value -r)
CPU=$(az keyvault secret show --name "polkadot-${prefix}-cpulimit" --vault-name "${key_vault_name}" | jq .value -r)
RAM=$(az keyvault secret show --name "polkadot-${prefix}-ramlimit" --vault-name "${key_vault_name}" | jq .value -r)
NODEKEY=$(az keyvault secret show --name "polkadot-${prefix}-nodekey" --vault-name "${key_vault_name}" | jq .value -r)

set +eE
trap - ERR
/usr/bin/docker stop polkadot
/usr/bin/docker rm polkadot
set -eE
trap default_trap ERR

/usr/bin/docker run --cpus $CPU --memory $${RAM}GB --kernel-memory $${RAM}GB --network=host --name polkadot --restart unless-stopped -d -p 127.0.0.1:9933:9933 -p 30333:30333 -v /data:/data:z chevdor/polkadot:latest polkadot --chain ${chain} --rpc-methods=Unsafe --pruning=archive

# Since polkadot container does not have curl inside - port it from host instance

exit_code=1
until [ $exit_code -eq 0 ]; do

  trap - ERR
  set +eE
  curl localhost:9933
  exit_code=$?
  set -eE
  trap default_trap ERR
  sleep 1

done

curl -o /usr/local/bin/install_consul.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/azure/install_consul.sh
curl -o /usr/local/bin/install_consulate.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/install_consulate.sh
curl -o /usr/local/bin/telegraf.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/azure/telegraf.sh

source /usr/local/bin/install_consul.sh
source /usr/local/bin/install_consulate.sh  
source /usr/local/bin/telegraf.sh ${prefix} $INSTANCE_ID
/usr/bin/systemctl restart telegraf

install_consul ${prefix} ${total_instance_count} ${lb-primary} ${lb-secondary} ${lb-tertiary}
install_consulate

# Join to the created cluster
cluster_members=0
until [ $cluster_members -gt ${total_instance_count} ]; do

  set +eE
  trap - ERR EXIT
  /usr/local/bin/consul join ${lb-primary} ${lb-secondary} ${lb-tertiary}
  trap default_trap ERR EXIT
  set -eE
  cluster_members=$(/usr/local/bin/consul members --status alive | wc -l)
  sleep 1s

done

curl -o /usr/local/bin/double-signing-control.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/double-signing-control.sh
curl -o /usr/local/bin/best-grep.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/best-grep.sh
curl -o /usr/local/bin/key-insert.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/azure-tests/init-helpers/azure/key-insert.sh
curl -o /usr/local/bin/watcher.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/azure-tests/init-helpers/azure/watcher.sh

chmod 700 /usr/local/bin/double-signing-control.sh
chmod 700 /usr/local/bin/best-grep.sh
chmod 700 /usr/local/bin/key-insert.sh
chmod 700 /usr/local/bin/watcher.sh

### This will add a crontab entry that will check nodes health from inside the VM and send data to the Azure Monitor
(echo '* * * * * /usr/local/bin/watcher.sh') | crontab -

# Create lock for the instance
n=0
set +eE
trap - ERR

until [ $n -ge 6 ]; do

  set -eE
  trap default_trap ERR
  node=$(curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_nodeRoles", "params":[]}' http://localhost:9933 | grep Full | wc -l)
  if [ "$node" != 1 ]; then 
    echo "ERROR! Node either does not work or work in not correct way"
    default_trap
  fi
  trap - ERR
  set +eE

  /usr/local/bin/consul lock prefix \
  "/usr/local/bin/double-signing-control.sh && /usr/local/bin/key-insert.sh '${key_vault_name}' '${prefix}' && consul kv delete blocks/.lock && (consul lock blocks \"while true; do /usr/local/bin/best-grep.sh; done\" &) && docker stop polkadot && docker rm polkadot && /usr/bin/docker run --cpus $${CPU} --memory $${RAM}GB --kernel-memory $${RAM}GB --network=host --name polkadot --restart unless-stopped -p 127.0.0.1:9933:9933 -p 30333:30333 -v /data:/data chevdor/polkadot:latest polkadot --chain ${chain} --validator --name '$NAME' --node-key '$NODEKEY'"

  /usr/bin/docker stop polkadot || true
  /usr/bin/docker rm polkadot || true
  /usr/bin/docker run --cpus $CPU --memory $${RAM}GB --kernel-memory $${RAM}GB --network=host --name polkadot --restart unless-stopped -d -p 127.0.0.1:9933:9933 -p 30333:30333 -v /data:/data:z chevdor/polkadot:latest polkadot --chain ${chain} --rpc-methods=Unsafe --pruning=archive
  pkill best-grep.sh
  sleep 10;
  n=$[$n+1]

done

default_trap
