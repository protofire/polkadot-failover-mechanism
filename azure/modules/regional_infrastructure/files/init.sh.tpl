#!/usr/bin/env bash

### Default trap for all errors in this script
default_trap ()
{
  echo "Catching error on line $LINENO. Shutting down instance";
  shutdown -P +1
  curl -i -XPOST 'http://localhost:12500/telegraf' --data-binary "health value=100000"
  /usr/local/bin/consul leave;
  docker stop polkadot
}

# Set verbose mode and quit on error flags
set -x -eE

hostname=$(hostname)
docker_name="polkadot"
data="/data"
polkadot_user_id=1000

# Install custom repos
curl -o /etc/yum.repos.d/azure-cli.repo -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/azure/azure-cli.repo
curl -o /etc/yum.repos.d/influxdb.repo -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/influxdb.repo

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
chown "$polkadot_user_id:$polkadot_user_id" /data -R
set -eE

trap default_trap ERR EXIT
# Mound disk and add automount entry in /etc/fstab
/usr/bin/grep -qxF "/dev/sdc /data xfs defaults 0 0" /etc/fstab || /usr/bin/echo "/dev/sdc /data xfs defaults 0 0" >> /etc/fstab

# Run docker with regular polkadot container inside of it
/usr/bin/systemctl enable docker
/usr/bin/systemctl start docker

# Fetch validator instance parameters from Secret Manager

trap - ERR
set +eE
set -o pipefail

exit_code=1
until [ $exit_code -eq 0 ]; do

  NAME=$(az keyvault secret show --name "polkadot-${prefix}-name" --vault-name "${key_vault_name}" | jq .value -r)
  exit_code=$?
  if [ $exit_code -ne 0 ]; then
    continue
  fi
  CPU=$(az keyvault secret show --name "polkadot-${prefix}-cpulimit" --vault-name "${key_vault_name}" | jq .value -r)
  exit_code=$?
  if [ $exit_code -ne 0 ]; then
    continue
  fi
  RAM=$(az keyvault secret show --name "polkadot-${prefix}-ramlimit" --vault-name "${key_vault_name}" | jq .value -r)
  exit_code=$?
  if [ $exit_code -ne 0 ]; then
    continue
  fi
  NODEKEY=$(az keyvault secret show --name "polkadot-${prefix}-nodekey" --vault-name "${key_vault_name}" | jq .value -r)
  exit_code=$?
  if [ $exit_code -ne 0 ]; then
    continue
  fi
  echo "Can not access secret at vault. Sleeping for 10s..."
  sleep 10

done

set +o pipefail
set -eE
trap default_trap ERR

curl -s -o /usr/local/bin/validator.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/validator.sh
source /usr/local/bin/validator.sh

start_polkadot_passive_mode "$docker_name" "$CPU" "$${RAM}GB" "${docker_image}" "${chain}" "$data" false \
                            "${expose_prometheus}" "${polkadot_prometheus_port}"

# Since polkadot container does not have curl inside - port it from host instance

exit_code=1
until [ $exit_code -eq 0 ]; do

  trap - ERR
  set +eE
  curl -X OPTIONS localhost:9933
  exit_code=$?
  set -eE
  trap default_trap ERR
  sleep 1

done

curl -o /usr/local/bin/install_consul.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/azure/install_consul.sh
curl -o /usr/local/bin/install_consulate.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/install_consulate.sh
curl -o /usr/local/bin/telegraf.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/azure/telegraf.sh

source /usr/local/bin/install_consul.sh
source /usr/local/bin/install_consulate.sh
source /usr/local/bin/telegraf.sh "${prefix}" "$hostname" "${group_name}" "${expose_prometheus}" "${polkadot_prometheus_port}" "${prometheus_port}"
/usr/bin/systemctl enable telegraf
/usr/bin/systemctl restart telegraf

install_consul "${prefix}" "${total_instance_count}" "${lb-primary}" "${lb-secondary}" "${lb-tertiary}"

install_consulate

lbs=()

[ -n "${lb-primary}" ] && lbs+=( "${lb-primary}" )
[ -n "${lb-secondary}" ] && lbs+=( "${lb-secondary}" )
[ -n "${lb-tertiary}" ] && lbs+=( "${lb-tertiary}" )

echo "LoadBalancers: $${lbs[@]}"

# Join to the created cluster
cluster_members=0
until [ $cluster_members -gt "${total_instance_count}" ]; do

  set +eE
  trap - ERR EXIT
  /usr/local/bin/consul join "$${lbs[@]}"
  trap default_trap ERR EXIT
  set -eE
  cluster_members=$(/usr/local/bin/consul members --status alive | wc -l)
  sleep 1s

done


curl -o /usr/local/bin/double-signing-control.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/double-signing-control.sh
curl -o /usr/local/bin/best-grep.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/best-grep.sh
curl -o /usr/local/bin/key-insert.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/azure/key-insert.sh
curl -o /usr/local/bin/watcher.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/watcher.sh

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
  node=$(curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_nodeRoles", "params":[]}' http://localhost:9933 | grep -c Full)
  if [ "$node" != 1 ]; then
    echo "ERROR! Node either does not work or work in not correct way"
    default_trap
  fi
  trap - ERR
  set +eE

  /usr/local/bin/consul lock prefix \
    "source /usr/local/bin/validator.sh && \
    /usr/local/bin/double-signing-control.sh && \
    start_polkadot_passive_mode $docker_name $CPU $${RAM}GB ${docker_image} ${chain} $data true ${expose_prometheus} ${polkadot_prometheus_port} && \
    /usr/local/bin/key-insert.sh '${key_vault_name}' '${prefix}' && \
    consul kv delete blocks/.lock && \
    (consul lock blocks \"while true; do /usr/local/bin/best-grep.sh; done\" &) && \
    start_polkadot_validator_mode $docker_name $CPU $${RAM}GB ${docker_image} ${chain} $data $NAME $NODEKEY ${expose_prometheus} ${polkadot_prometheus_port}"

  start_polkadot_passive_mode "$docker_name" "$CPU" "$${RAM}GB" "${docker_image}" "${chain}" "$data" false \
                              "${expose_prometheus}" "${polkadot_prometheus_port}"
  pkill best-grep.sh
  sleep 10;
  n=$((n+1))

done

default_trap
