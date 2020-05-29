#!/usr/bin/env bash

# Set quit on error flags
set -x -eE

### Function that attaches disk
disk_attach ()
{
  # Loop five times
  n=0
  until [ $n -ge 5 ]; do

    # Check that there is a disk available
    VOLUME=$(aws ec2 describe-volumes --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --filters "Name=status,Values=available" "Name=tag:prefix,Values=${prefix}" --query 'Volumes[*].VolumeId' --output json | jq .[0] -r)
    # Attach volume as /dev/sdb and allow errors. If attach successful - quit from the loop
    set +e
    aws ec2 attach-volume --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --device /dev/sdb --instance-id $(curl -s http://169.254.169.254/latest/meta-data/instance-id) --volume-id $VOLUME && break
    set -e
    # If attach was not successfull - increment over loop.
    n=$[$n+1]
    sleep $[ ( $RANDOM % 10 )  + 1 ]s
    
    # If we reached the fifth retry - fail script, send alarm to cloudwatch and shutdown instance
    if [ $n -eq 5 ]; then
      echo "ERROR! No disk can be attached. Shutting down instance"
      aws cloudwatch put-metric-data --region ${primary-region} --metric-name 'Health report' --dimensions AutoScalingGroupName=${autoscaling-name} --namespace '${prefix}' --value 1000;
      aws cloudwatch put-metric-data --region ${secondary-region} --metric-name 'Health report' --dimensions AutoScalingGroupName=${autoscaling-name} --namespace '${prefix}' --value 1000; 
      aws cloudwatch put-metric-data --region ${tertiary-region} --metric-name 'Health report' --dimensions AutoScalingGroupName=${autoscaling-name} --namespace '${prefix}' --value 1000;
    fi
  done

  # Check that the instance is attached because attach curl request responds before the attachment is actually complete
  n=0
  VOLUMES=0
  until [ $VOLUMES -gt 0 ]; do

    VOLUMES=$(aws ec2 describe-volumes --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --filters "Name=status,Values=in-use" --volume-ids $VOLUME --query 'Volumes[*].VolumeId' --output json | jq '. | length')
    sleep 10
  done

  # Optionally marks data disk with "delete on termination"
  if $(${delete_on_termination}); then
    aws ec2 modify-instance-attribute --instance-id $(curl -s http://169.254.169.254/latest/meta-data/instance-id) --block-device-mappings "[{\"DeviceName\": \"/dev/sdb\",\"Ebs\":{\"DeleteOnTermination\":true}}]"
  fi
}
### End of function

### Start of main script
# Install unzip, docker, jq, git, awscli
/usr/bin/yum update -y
/usr/bin/yum install unzip docker jq git -y
/usr/bin/curl -s "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
/usr/bin/unzip -qq -o awscliv2.zip
./aws/install --update

# Check that instance profile is attached
until aws sts get-caller-identity; do

  echo "No AWS credentials found. Retrying..."
  sleep 5

done  

default_trap () 
{
  echo \"Catching error on line $LINENO. Shutting down instance\";
  aws cloudwatch put-metric-data --region ${primary-region} --metric-name 'Health report' --dimensions AutoScalingGroupName=${autoscaling-name} --namespace '${prefix}' --value 1000; 
  aws cloudwatch put-metric-data --region ${secondary-region} --metric-name 'Health report' --dimensions AutoScalingGroupName=${autoscaling-name} --namespace '${prefix}' --value 1000; 
  aws cloudwatch put-metric-data --region ${tertiary-region} --metric-name 'Health report' --dimensions AutoScalingGroupName=${autoscaling-name} --namespace '${prefix}' --value 1000; 
  /usr/local/bin/consul leave;
  docker stop polkadot
  shutdown -P +1  
}

# On error notify cloudwatch and shutdown instance
  trap default_trap ERR EXIT

# Check that there is no shutting down instance to prevent bug when disk is still not deattached from previous instance before attaching to this one
INSTANCES_COUNT=$(aws ec2 describe-instances --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --filters "Name=instance-state-name,Values=shutting-down,stopping" "Name=tag:prefix,Values=${prefix}" --query 'Reservations[*].Instances[*].InstanceId' --output json | jq '. | length')

until [ $INSTANCES_COUNT -eq 0 ]; do

  sleep 5;
  INSTANCES_COUNT=$(aws ec2 describe-instances --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --filters "Name=instance-state-name,Values=shutting-down,stopping" "Name=tag:prefix,Values=${prefix}" --query 'Reservations[*].Instances[*].InstanceId' --output json | jq '. | length')

done

# Check if there are data disks available
VOLUMES=$(aws ec2 describe-volumes --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --filters "Name=status,Values=available" "Name=tag:prefix,Values=${prefix}" --query 'Volumes[*].VolumeId' --output json)
VOLUMES_COUNT=$(echo $VOLUMES | jq '. | length')

# It there are available data disks - attach one of them
if [ $VOLUMES_COUNT -gt 0 ]; then

  disk_attach

# If not - create a new one and attach
else

  VOLUME=$(aws ec2 create-volume --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --availability-zone $(curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone) --size ${disk_size} --tag-specifications "ResourceType=volume,Tags=[{Key=prefix,Value=${prefix}},{Key=Name,Value=${prefix}-polkadot-failover-data}]" | jq .VolumeId -r)

  until [ $VOLUMES_COUNT -gt 0 ]; do

    VOLUMES=$(aws ec2 describe-volumes --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --filters "Name=status,Values=available" "Name=tag:prefix,Values=${prefix}" --query 'Volumes[*].VolumeId' --output json)
    VOLUMES_COUNT=$(echo $VOLUMES | jq '. | length')
  
    if [ $VOLUMES_COUNT -gt 0 ]; then

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

/usr/bin/docker run --cpus ${cpu_limit} --memory ${ram_limit}GB --kernel-memory ${ram_limit}GB --network=host --name polkadot --restart unless-stopped -d -p 127.0.0.1:9933:9933 -p 30333:30333 -v /data:/data chevdor/polkadot:latest polkadot --chain ${chain} --rpc-methods=Unsafe --pruning=archive

exit_code=1
set +eE
trap - ERR EXIT

until [ $exit_code -eq 0 ]; do

  curl localhost:9933
  exit_code=$?
  sleep 1

done

set -eE
trap default_trap ERR EXIT

# Clone and install consul
git clone https://github.com/hashicorp/terraform-aws-consul.git
terraform-aws-consul/modules/install-consul/install-consul --version 1.7.2

## Override consul config. Add join through load balancers and leave_on_terminate flag

cat <<EOF > /opt/consul/config/new-retry-join.json
{
  "retry_join": [
    "${lb-primary}", "${lb-secondary}", "${lb-tertiary}"
  ],
  "leave_on_terminate": true,
  "bootstrap_expect": ${total_instance_count}
}
EOF

# Start consul with a new script in server mode
/opt/consul/bin/run-consul --server --datacenter "${prefix}"
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

# Fetch validator instance name from SSM
NAME=$(aws ssm get-parameter --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --name "/polkadot/validator-failover/${prefix}/name" | jq -r .Parameter.Value)
CPU=$(aws ssm get-parameter --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --name "/polkadot/validator-failover/${prefix}/cpu_limit" | jq -r .Parameter.Value)
RAM=$(aws ssm get-parameter --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --name "/polkadot/validator-failover/${prefix}/ram_limit" | jq -r .Parameter.Value)
NODEKEY=$(aws ssm get-parameter --region $(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region) --name "/polkadot/validator-failover/${prefix}/node_key" | jq -r .Parameter.Value)

curl -o /usr/local/bin/double-signing-control.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/double-signing-control.sh
curl -o /usr/local/bin/best-grep.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/best-grep.sh
curl -o /usr/local/bin/key-insert.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/aws/key-insert.sh
curl -o /usr/local/bin/watcher.sh -L https://raw.githubusercontent.com/protofire/polkadot-failover-mechanism/master/init-helpers/aws/watcher.sh

chmod 700 /usr/local/bin/double-signing-control.sh
chmod 700 /usr/local/bin/best-grep.sh
chmod 700 /usr/local/bin/key-insert.sh
chmod 700 /usr/local/bin/watcher.sh

### This will add a crontab entry that will check nodes health from inside the VM and send data to the CloudWatch
(echo '* * * * * /usr/local/bin/watcher.sh ${prefix} ${autoscaling-name} ${primary-region} ${secondary-region} ${tertiary-region}') | crontab -

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

  /usr/local/bin/consul lock prefix "/usr/local/bin/double-signing-control.sh && /usr/local/bin/key-insert.sh ${prefix} && (consul kv delete blocks/.lock && consul lock blocks \"while true; do /usr/local/bin/best-grep.sh; done\" &) && docker stop polkadot && docker rm polkadot && /usr/bin/docker run --cpus $${CPU} --memory $${RAM}GB --kernel-memory $${RAM}GB --network=host --name polkadot --restart unless-stopped -p 127.0.0.1:9933:9933 -p 30333:30333 -v /data:/data chevdor/polkadot:latest polkadot --chain ${chain} --validator --name '$NAME' --node-key '$NODEKEY'"

  /usr/bin/docker stop polkadot || true
  /usr/bin/docker rm polkadot || true
  /usr/bin/docker run --cpus $CPU --memory $${RAM}GB --kernel-memory $${RAM}GB --network=host --name polkadot --restart unless-stopped -d -p 127.0.0.1:9933:9933 -p 30333:30333 -v /data:/data:z chevdor/polkadot:latest polkadot --chain ${chain} --rpc-methods=Unsafe --pruning=archive
  pkill best-grep.sh

  sleep 10;
  n=$[$n+1]

done

default_trap
