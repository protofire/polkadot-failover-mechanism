#!/usr/bin/env bash

default_trap ()
{
  shutdown -P +1
  /usr/local/bin/consul leave
  docker stop polkadot
}

# Set verbose mode and quit on error flags
set -x -eE

# Install unzip, docker, jq, git, awscli
/usr/bin/yum update -y
/usr/bin/yum install unzip docker jq git -y

# Installing Stackdriver
curl -sSO https://dl.google.com/cloudagents/add-monitoring-agent-repo.sh
bash add-monitoring-agent-repo.sh
yum install -y stackdriver-agent-6.*
curl -o /opt/stackdriver/collectd/etc/collectd.d/statsd.conf -L https://raw.githubusercontent.com/Stackdriver/stackdriver-agent-service-configs/master/etc/collectd.d/statsd.conf
service stackdriver-agent restart

# get helper functons to install and configure necessary tools and systems
curl -o /usr/local/bin/install_consul.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/gcp/install_consul.sh
curl -o /usr/local/bin/install_consulate.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/install_consulate.sh
curl -o /usr/local/bin/stackdriver.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/gcp/stackdriver.sh

source /usr/local/bin/install_consul.sh
source /usr/local/bin/install_consulate.sh  
source /usr/local/bin/stackdriver.sh

install_consul ${prefix} ${total_instance_count}
install_consulate
generate_collectd_checks
generate_collectd_rewrite

# Apply new collectd config
service stackdriver-agent restart

# Verify consul status
cluster_members=0
until [ $cluster_members -gt ${total_instance_count} ]; do

  cluster_members=$(/usr/local/bin/consul members --status alive | wc -l)
  sleep 10s

done

# Leave cluster if any errors
trap default_trap ERR EXIT

# Make file system on a disk without force. Allow error. Error will be thrown if the disk already has filesystem. This will prevent data erasing.
set +eE
trap - ERR
/usr/sbin/mkfs.xfs /dev/sdb
trap default_trap ERR
set -eE
# Mound disk and add automount entry in /etc/fstab
/usr/bin/mkdir /data
/usr/bin/mount /dev/sdb /data
/usr/bin/echo "/dev/sdb /data xfs defaults 0 0" >> /etc/fstab

chown 1000:1000 /data

# Run docker with regular polkadot container inside of it
/usr/bin/systemctl start docker

# Fetch validator instance parameters from Secret Manager
NAME_NAME=$(gcloud secrets list --format=json --filter="name ~ ${prefix}_name AND labels.prefix=${prefix} AND labels.type != key" | jq -r .[0].name)
NAME=$(gcloud secrets versions access $(gcloud secrets versions list $NAME_NAME --format json | jq '.[] | select(.state == "ENABLED") | .name' -r) --secret=$NAME_NAME --format json | jq .payload.data -r | base64 -d)
CPU_NAME=$(gcloud secrets list --format=json --filter="name ~ ${prefix}_cpulimit AND labels.prefix=${prefix} AND labels.type != key" | jq -r .[0].name)
CPU=$(gcloud secrets versions access $(gcloud secrets versions list $CPU_NAME --format json | jq '.[] | select(.state == "ENABLED") | .name' -r) --secret=$CPU_NAME --format json | jq .payload.data -r | base64 -d)
RAM_NAME=$(gcloud secrets list --format=json --filter="name ~ ${prefix}_ramlimit AND labels.prefix=${prefix} AND labels.type != key" | jq -r .[0].name)
RAM=$(gcloud secrets versions access $(gcloud secrets versions list $RAM_NAME --format json | jq '.[] | select(.state == "ENABLED") | .name' -r) --secret=$RAM_NAME --format json | jq .payload.data -r | base64 -d)
NODEKEY_NAME=$(gcloud secrets list --format=json --filter="name ~ ${prefix}_nodekey AND labels.prefix=${prefix} AND labels.type != key" | jq -r .[0].name)
NODEKEY=$(gcloud secrets versions access $(gcloud secrets versions list $NODEKEY_NAME --format json | jq '.[] | select(.state == "ENABLED") | .name' -r) --secret=$NODEKEY_NAME --format json | jq .payload.data -r | base64 -d)

/usr/bin/docker run --cpus $CPU --memory $${RAM}GB --kernel-memory $${RAM}GB --network=host --name polkadot --restart unless-stopped -d -p 127.0.0.1:9933:9933 -p 30333:30333 -v /data:/data:z chevdor/polkadot:latest polkadot --chain ${chain} --rpc-methods=Unsafe --pruning=archive

exit_code=1
until [ $exit_code -eq 0 ]; do

  set +eE
  trap - ERR
  curl localhost:9933
  exit_code=$?
  set -Ee
  trap default_trap ERR
  sleep 1

done

curl -o /usr/local/bin/double-signing-control.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/double-signing-control.sh
curl -o /usr/local/bin/best-grep.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/best-grep.sh
curl -o /usr/local/bin/key-insert.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/gcp/key-insert.sh

chmod 700 /usr/local/bin/double-signing-control.sh
chmod 700 /usr/local/bin/best-grep.sh
chmod 700 /usr/local/bin/key-insert.sh

# Create lock for the instance
n=0

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
  
  /usr/local/bin/consul lock prefix "/usr/local/bin/double-signing-control.sh && /usr/local/bin/key-insert.sh ${prefix} && consul kv delete blocks/.lock && (consul lock blocks \"while true; do /usr/local/bin/best-grep.sh; done\" &) && docker stop polkadot && docker rm polkadot && /usr/bin/docker run --cpus $${CPU} --memory $${RAM}GB --kernel-memory $${RAM}GB --network=host --name polkadot --restart unless-stopped -p 127.0.0.1:9933:9933 -p 30333:30333 -v /data:/data chevdor/polkadot:latest polkadot --chain ${chain} --validator --name '$NAME' --node-key '$NODEKEY'"

  /usr/bin/docker stop polkadot || true
  /usr/bin/docker rm polkadot || true
  /usr/bin/docker run --cpus $CPU --memory $${RAM}GB --kernel-memory $${RAM}GB --network=host --name polkadot --restart unless-stopped -d -p 127.0.0.1:9933:9933 -p 30333:30333 -v /data:/data:z chevdor/polkadot:latest polkadot --chain ${chain} --rpc-methods=Unsafe --pruning=archive
  pkill best-grep.sh

  sleep 10;
  n=$[$n+1]

done

default_trap
