#!/usr/bin/env bash

# Set quit on error flags
set -x -eE

set_health_metrics() {
      local metric_name=$1
      local value=$2
      curl -i -XPOST 'http://localhost:12500/telegraf' --data-binary "$metric_name value=$value"
}

### Function that attaches disk
disk_attach ()
{
  # Loop five times
  n=0
  until [ $n -ge 5 ]; do

    # Check that there is a disk available
    VOLUME=$(aws ec2 describe-volumes --region "$region" --filters "Name=status,Values=available" "Name=tag:prefix,Values=${prefix}" --query 'Volumes[*].VolumeId' --output json | jq .[0] -r)
    # Attach volume as /dev/sdb and allow errors. If attach successful - quit from the loop
    set +e
    aws ec2 attach-volume --region "$region" --device /dev/sdb --instance-id "$instance_id" --volume-id "$VOLUME" && break
    set -e
    # If attach was not successful - increment over loop.
    n=$((n + 1))
    sleep $((( RANDOM % 10 ) + 1 ))s
    
    # If we reached the fifth retry - fail script, send alarm to cloudwatch and shutdown instance
    if [ $n -eq 5 ]; then
      echo "ERROR! No disk can be attached. Shutting down instance"
      set_health_metrics "$health_metric_name" 1000
    fi
  done

  # Check that the instance is attached because attach curl request responds before the attachment is actually complete
  n=0
  VOLUMES=0
  until [ $VOLUMES -gt 0 ]; do
    VOLUMES=$(aws ec2 describe-volumes --region "$region" --filters "Name=status,Values=in-use" --volume-ids "$VOLUME" --query 'Volumes[*].VolumeId' --output json | jq '. | length')
    sleep 10
  done

  # Optionally marks data disk with "delete on termination"
  if "${delete_on_termination}"; then
    aws ec2 modify-instance-attribute --instance-id "$instance_id" --block-device-mappings "[{\"DeviceName\": \"/dev/sdb\",\"Ebs\":{\"DeleteOnTermination\":true}}]"
  fi
}
### End of function

### Start of main script
# Install unzip, docker, jq, git, awscli
curl -o /etc/yum.repos.d/influxdb.repo -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/aws/influxdb.repo
/usr/bin/yum update -y
/usr/bin/yum install unzip docker jq git telegraf nc -y
/usr/bin/curl -s "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
/usr/bin/unzip -qq -o awscliv2.zip
./aws/install --update

region=$(curl -s http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region)
instance_id=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
zone=$(curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone)

# Get hostname
INSTANCE_ID=$(hostname)

lbs=()
[ -n "${lb-primary}" ] && lbs+=( "${lb-primary}" )
[ -n "${lb-secondary}" ] && lbs+=( "${lb-secondary}" )
[ -n "${lb-tertiary}" ] && lbs+=( "${lb-tertiary}" )

regions=( "${primary-region}" "${secondary-region}" "${tertiary-region}" )
health_metric_name="health"

# Check that instance profile is attached
until aws sts get-caller-identity; do

  echo "No AWS credentials found. Retrying..."
  sleep 5

done  

default_trap () 
{
  echo \"Catching error on line $LINENO. Shutting down instance\";
  set_health_metrics "$health_metric_name" 1000
  /usr/local/bin/consul leave;
  docker stop polkadot
  shutdown -P +1  
}

# On error notify cloudwatch and shutdown instance
  trap default_trap ERR EXIT

# Check that there is no shutting down instance to prevent bug when disk is still not de-attached from previous instance before attaching to this one
INSTANCES_COUNT=$(aws ec2 describe-instances --region "$region" --filters "Name=instance-state-name,Values=shutting-down,stopping" "Name=tag:prefix,Values=${prefix}" --query 'Reservations[*].Instances[*].InstanceId' --output json | jq '. | length')

until [ "$INSTANCES_COUNT" -eq 0 ]; do

  sleep 5;
  INSTANCES_COUNT=$(aws ec2 describe-instances --region "$region" --filters "Name=instance-state-name,Values=shutting-down,stopping" "Name=tag:prefix,Values=${prefix}" --query 'Reservations[*].Instances[*].InstanceId' --output json | jq '. | length')

done

# Check if there are data disks available
VOLUMES=$(aws ec2 describe-volumes --region "$region" --filters "Name=status,Values=available" "Name=tag:prefix,Values=${prefix}" --query 'Volumes[*].VolumeId' --output json)
VOLUMES_COUNT=$(echo "$VOLUMES" | jq '. | length')

# It there are available data disks - attach one of them
if [ "$VOLUMES_COUNT" -gt 0 ]; then

  disk_attach

# If not - create a new one and attach
else

  VOLUME=$(aws ec2 create-volume --region "$region" --availability-zone "$zone" --size "${disk_size}" --tag-specifications "ResourceType=volume,Tags=[{Key=prefix,Value=${prefix}},{Key=Name,Value=${prefix}-polkadot-failover-data}]" | jq .VolumeId -r)

  until [ "$VOLUMES_COUNT" -gt 0 ]; do

    VOLUMES=$(aws ec2 describe-volumes --region "$region" --filters "Name=status,Values=available" "Name=tag:prefix,Values=${prefix}" --query 'Volumes[*].VolumeId' --output json)
    VOLUMES_COUNT=$(echo "$VOLUMES" | jq '. | length')
  
    if [ "$VOLUMES_COUNT" -gt 0 ]; then

      disk_attach
      break

    else

      sleep 5

    fi

  done

fi

flag=0
until [ $flag -gt 0 ]; do
# Loop through attached disks
for DISK in /dev/nvme?; do

    LABEL=$(/sbin/ebsnvme-id -b $${DISK})
    # Find the one that is attached as /dev/sdb
    if [ "$${LABEL}" == "sdb" ] || [ "$${LABEL}" == "/dev/sdb" ]; then

        # Make file system on a disk without force. Allow error. Error will be thrown if the disk already has filesystem. This will prevent data erasing.
        /usr/sbin/mkfs.xfs $${DISK}n1 || true
        # Mound disk and add automount entry in /etc/fstab
        UUID=$(/usr/bin/lsblk $${DISK}n1 -nr -o UUID)
        /usr/bin/mkdir /data
        /usr/bin/mount $${DISK}n1 /data
        /usr/bin/echo "UUID=$${UUID} /data xfs defaults 0 0" >> /etc/fstab
        # Change owner of data folder to ec2-user
        /usr/bin/chown ec2-user:ec2-user /data
        # Remove keys inside of data folder TODO
        rm -R /data/chains/*/keystore || true
    else
      flag=1
    fi

done
done

# Run docker with regular polkadot container inside of it
/usr/bin/systemctl start docker

/usr/bin/docker run \
  --cpus "${cpu_limit}" \
  --memory "${ram_limit}GB" \
  --kernel-memory "${ram_limit}GB" \
  --network=host \
  --name polkadot \
  --restart unless-stopped \
  -d \
  -p 127.0.0.1:9933:9933 \
  -p 30333:30333 \
  -v /data:/data \
  "${docker_image}" --chain "${chain}" --rpc-methods=Unsafe --rpc-external --pruning=archive -d /data

exit_code=1
set +eE
trap - ERR EXIT

until [ $exit_code -eq 0 ]; do

  curl -X OPTIONS localhost:9933
  exit_code=$?
  sleep 1

done

curl -o /usr/local/bin/install_consul.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/aws/install_consul.sh
curl -o /usr/local/bin/install_consulate.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/install_consulate.sh
curl -o /usr/local/bin/telegraf.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/aws/telegraf.sh

source /usr/local/bin/install_consul.sh
source /usr/local/bin/install_consulate.sh
source /usr/local/bin/telegraf.sh "${prefix}" "$INSTANCE_ID" "$region" "${autoscaling-name}" "$instance_id"
/usr/bin/systemctl enable telegraf
/usr/bin/systemctl restart telegraf

install_consul "${prefix}" "${total_instance_count}" "${lb-primary}" "${lb-secondary}" "${lb-tertiary}"
install_consulate

set -eE
trap default_trap ERR EXIT

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

# Fetch validator instance name from SSM
NAME=$(aws ssm get-parameter --region "$region" --name "/polkadot/validator-failover/${prefix}/name" | jq -r .Parameter.Value)
CPU=$(aws ssm get-parameter --region "$region" --name "/polkadot/validator-failover/${prefix}/cpu_limit" | jq -r .Parameter.Value)
RAM=$(aws ssm get-parameter --region "$region" --name "/polkadot/validator-failover/${prefix}/ram_limit" | jq -r .Parameter.Value)
NODEKEY=$(aws ssm get-parameter --region "$region" --name "/polkadot/validator-failover/${prefix}/node_key" | jq -r .Parameter.Value)

curl -o /usr/local/bin/double-signing-control.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/double-signing-control.sh
curl -o /usr/local/bin/best-grep.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/best-grep.sh
curl -o /usr/local/bin/key-insert.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/aws/key-insert.sh
curl -o /usr/local/bin/watcher.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/dev/init-helpers/watcher.sh

chmod 700 /usr/local/bin/double-signing-control.sh
chmod 700 /usr/local/bin/best-grep.sh
chmod 700 /usr/local/bin/key-insert.sh
chmod 700 /usr/local/bin/watcher.sh

### This will add a crontab entry that will check nodes health from inside the VM and send data to the CloudWatch
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
    "/usr/local/bin/double-signing-control.sh && \
    /usr/local/bin/key-insert.sh ${prefix} && \
    (consul kv delete blocks/.lock && \
    consul lock blocks \"while true; do /usr/local/bin/best-grep.sh; done\" &) && \
    docker stop polkadot && \
    docker rm polkadot && \
    /usr/bin/docker run \
    --cpus $${CPU} \
    --memory $${RAM}GB \
    --kernel-memory $${RAM}GB \
    --network=host \
    --name polkadot \
    --restart unless-stopped \
    -p 127.0.0.1:9933:9933 \
    -p 30333:30333 \
    -v /data:/data:z \
    ${docker_image} --chain ${chain} --validator --name '$NAME' --node-key '$NODEKEY' -d /data"

  /usr/bin/docker stop polkadot || true
  /usr/bin/docker rm polkadot || true
  /usr/bin/docker run \
  --cpus "$CPU" \
  --memory $${RAM}GB \
  --kernel-memory $${RAM}GB \
  --network=host \
  --name polkadot \
  --restart unless-stopped \
  -d \
  -p 127.0.0.1:9933:9933 \
  -p 30333:30333 \
  -v /data:/data:z \
  "${docker_image}" --chain "${chain}" --rpc-methods=Unsafe --rpc-external --pruning=archive -d /data
  pkill best-grep.sh

  sleep 10;
  n=$((n+1))

done

default_trap
