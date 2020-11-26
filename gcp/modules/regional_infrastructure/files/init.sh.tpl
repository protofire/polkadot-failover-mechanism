#!/usr/bin/env bash

default_trap ()
{
  echo "Catching error on line $LINENO. Shutting down instance";
  shutdown -P +1
  curl -i -XPOST 'http://localhost:12500/telegraf' --data-binary "health value=100000"
  /usr/local/bin/consul leave
  docker stop polkadot
}

# Set verbose mode and quit on error flags
set -x -eE

hostname=$(hostname)
docker_name="polkadot"
data="/data"
polkadot_user_id=1000

curl -o /etc/yum.repos.d/influxdb.repo -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/influxdb.repo

# Install unzip, docker, jq, git
/usr/bin/yum update -y
/usr/bin/yum install telegraf unzip docker jq git nc -y

zone=$(curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/zone)

# get helper functions to install and configure necessary tools and systems
curl -o /usr/local/bin/install_consul.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/gcp/install_consul.sh
curl -o /usr/local/bin/install_consulate.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/install_consulate.sh
curl -o /usr/local/bin/telegraf.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/gcp/telegraf.sh
curl -s -o /usr/local/bin/validator.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/validator.sh

source /usr/local/bin/validator.sh
source /usr/local/bin/install_consul.sh
source /usr/local/bin/install_consulate.sh
source /usr/local/bin/telegraf.sh "${prefix}" "$hostname" "${project}" "${metrics_namespace}" "$${zone##*/}" "${group_name}" \
                                  "${expose_prometheus}" "${polkadot_prometheus_port}" "${prometheus_port}"
/usr/bin/systemctl enable telegraf
/usr/bin/systemctl restart telegraf

install_consul "${prefix}" "${total_instance_count}"
install_consulate

# Verify consul status
cluster_members=0
until [ $cluster_members -gt "${total_instance_count}" ]; do
  cluster_members=$(/usr/local/bin/consul members --status alive | wc -l)
  sleep 10s
done

# Leave cluster if any errors
trap default_trap ERR EXIT

# Make file system on a disk without force. Allow error. Error will be thrown if the disk already has filesystem. This will prevent data erasing.
set +eE
trap - ERR
/usr/sbin/mkfs.xfs /dev/sdb
/usr/bin/mkdir -p /data
if grep -qs '/data ' /proc/mounts; then
    echo "Disk already mounted, skipping..."
else
    mount /dev/sdb /data
fi
grep /dev/sdb /etc/fstab || /usr/bin/echo "/dev/sdb /data xfs defaults 0 0" >> /etc/fstab
set -eE
trap default_trap ERR

chown "$polkadot_user_id:$polkadot_user_id" /data -R

# Run docker with regular polkadot container inside of it
/bin/systemctl enable docker
/bin/systemctl start docker

# Fetch validator instance parameters from Secret Manager
NAME_NAME=$(gcloud secrets list --format=json --filter="name ~ ${prefix}_name AND labels.prefix=${prefix} AND labels.type != key" | jq -r .[0].name)
NAME_NAME_ACCESS=$(gcloud secrets versions list "$NAME_NAME" --format json | jq '.[] | select(.state == "ENABLED") | .name' -r)
NAME=$(gcloud secrets versions access "$NAME_NAME_ACCESS" --secret="$NAME_NAME" --format json | jq .payload.data -r | base64 -d)

CPU_NAME=$(gcloud secrets list --format=json --filter="name ~ ${prefix}_cpulimit AND labels.prefix=${prefix} AND labels.type != key" | jq -r .[0].name)
CPU_NAME_ACCESS=$(gcloud secrets versions list "$CPU_NAME" --format json | jq '.[] | select(.state == "ENABLED") | .name' -r)
CPU=$(gcloud secrets versions access "$CPU_NAME_ACCESS" --secret="$CPU_NAME" --format json | jq .payload.data -r | base64 -d)

RAM_NAME=$(gcloud secrets list --format=json --filter="name ~ ${prefix}_ramlimit AND labels.prefix=${prefix} AND labels.type != key" | jq -r .[0].name)
RAM_NAME_ACCESS=$(gcloud secrets versions list "$RAM_NAME" --format json | jq '.[] | select(.state == "ENABLED") | .name' -r)
RAM=$(gcloud secrets versions access "$RAM_NAME_ACCESS" --secret="$RAM_NAME" --format json | jq .payload.data -r | base64 -d)

NODEKEY_NAME=$(gcloud secrets list --format=json --filter="name ~ ${prefix}_nodekey AND labels.prefix=${prefix} AND labels.type != key" | jq -r .[0].name)
NODEKEY_NAME_ACCESS=$(gcloud secrets versions list "$NODEKEY_NAME" --format json | jq '.[] | select(.state == "ENABLED") | .name' -r)
NODEKEY=$(gcloud secrets versions access "$NODEKEY_NAME_ACCESS" --secret="$NODEKEY_NAME" --format json | jq .payload.data -r | base64 -d)

start_polkadot_passive_mode "$docker_name" "$CPU" "$${RAM}GB" "${docker_image}" "${chain}" "$data" false \
                            "${expose_prometheus}" "${polkadot_prometheus_port}"

exit_code=1
until [ $exit_code -eq 0 ]; do

  set +eE
  trap - ERR
  curl -X OPTIONS localhost:9933
  exit_code=$?
  set -Ee
  trap default_trap ERR
  sleep 1

done

curl -o /usr/local/bin/double-signing-control.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/double-signing-control.sh
curl -o /usr/local/bin/best-grep.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/best-grep.sh
curl -o /usr/local/bin/key-insert.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/gcp/key-insert.sh
curl -o /usr/local/bin/watcher.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/watcher.sh

chmod 700 /usr/local/bin/double-signing-control.sh
chmod 700 /usr/local/bin/best-grep.sh
chmod 700 /usr/local/bin/key-insert.sh
chmod 700 /usr/local/bin/watcher.sh

### This will add a crontab entry that will check nodes health from inside the VM and send data to the GCP Monitor
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
    "source /usr/local/bin/validator.sh && \
    /usr/local/bin/double-signing-control.sh && \
    start_polkadot_passive_mode $docker_name $CPU $${RAM}GB ${docker_image} ${chain} $data true ${expose_prometheus} ${polkadot_prometheus_port} && \
    /usr/local/bin/key-insert.sh ${prefix} && \
    (consul kv delete blocks/.lock && \
    consul lock blocks \"while true; do /usr/local/bin/best-grep.sh; done\" &) && \
    start_polkadot_validator_mode $docker_name $CPU $${RAM}GB ${docker_image} ${chain} $data $NAME $NODEKEY ${expose_prometheus} ${polkadot_prometheus_port}"

  start_polkadot_passive_mode "$docker_name" "$CPU" "$${RAM}GB" "${docker_image}" "${chain}" "$data" false "${expose_prometheus}" "${polkadot_prometheus_port}"
  pkill best-grep.sh
  sleep 10;
  n=$((n+1))

done

default_trap
