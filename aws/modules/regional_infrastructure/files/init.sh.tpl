#!/usr/bin/env bash

# Set quit on error flags
set -x -e -E

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
    fi
done

usermod -a -G docker ec2-user
# Run docker with regular polkadot container inside of it
/usr/bin/systemctl start docker

/usr/bin/docker run --cpus ${cpu_limit} --memory ${ram_limit}GB --kernel-memory ${ram_limit}GB --name polkadot --restart unless-stopped -d -p 30333:30333 -p 127.0.0.1:9933:9933 -v /data:/data chevdor/polkadot:latest polkadot --chain ${chain} --rpc-external --rpc-cors=all --pruning=archive

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

cat <<EOF >/usr/local/bin/watcher.sh
#!/bin/bash
\$(docker inspect -f "{{.State.Running}}" polkadot && curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_health", "params":[]}' http://127.0.0.1:9933);
STATE=\$?
BLOCK_NUMBER=\$((\$(curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "chain_getBlock", "params":[]}' http://127.0.0.1:9933 | jq .result.block.header.number -r)))
AMIVALIDATOR=\$(curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_nodeRoles", "params":[]}' http://127.0.0.1:9933 | jq -r .result[0])
if [ "\$AMIVALIDATOR" == "Authority" ]; then
  AMIVALIDATOR=1
else
  AMIVALIDATOR=0
fi

regions=( ${primary-region} ${secondary-region} ${tertiary-region} )

for i in "\$${regions[@]}"; do
  aws cloudwatch put-metric-data --region \$i --metric-name "Health report" --dimensions AutoScalingGroupName=${autoscaling-name} --namespace "${prefix}" --value "\$STATE"
  aws cloudwatch put-metric-data --region \$i --metric-name "Block Number" --dimensions InstanceID="\$(curl --silent http://169.254.169.254/latest/meta-data/instance-id)" --namespace "${prefix}" --value "\$BLOCK_NUMBER"
  aws cloudwatch put-metric-data --region \$i --metric-name "Validator count" --dimensions AutoScalingGroupName=${autoscaling-name} --namespace "${prefix}" --value "\$AMIVALIDATOR"
done

EOF

chmod 777 /usr/local/bin/watcher.sh

### This will add a crontab entry that will check nodes health from inside the VM and send data to the CloudWatch
(echo '* * * * * /usr/local/bin/watcher.sh') | crontab -

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

cat <<EOF >/usr/local/bin/double-signing-control.sh
#!/bin/bash

set -x

BEST=\$(/usr/local/bin/consul kv get best_block)
retVal=\$?

set -eE
echo "Previous validator best block - \$BEST"

if [ "\$retVal" -eq 0 ]; then

  VALIDATED=0
  until [ "\$VALIDATED" -gt "\$BEST" ]; do

    BEST_TEMP=\$(/usr/local/bin/consul kv get best_block)
    if [ "\$BEST_TEMP" != "\$BEST" ]; then
      consul leave
      shutdown now
      exit 1
    else
      BEST=\$BEST_TEMP
      VALIDATED=\$(/usr/bin/docker logs polkadot 2>&1 | /usr/bin/grep finalized | /usr/bin/tail -n 1)
      VALIDATED=\$(/usr/bin/echo \$${VALIDATED##*#} | /usr/bin/cut -d'(' -f1 | /usr/bin/xargs)
      echo "Previous validator best block - \$BEST, new validator validated block - \$VALIDATED"
      sleep 10
    fi

  done
fi

EOF

cat <<EOF >/usr/local/bin/key-insert.sh
#!/bin/bash

set -x -e -E

region="\$(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region)"
key_names=\$(aws ssm get-parameters-by-path --region \$region --recursive --path /polkadot/validator-failover/${prefix}/keys/ | jq .Parameters[].Name | awk -F'/' '{print\$(NF-1)}' | sort | uniq)

for key_name in \$${key_names[@]} ; do

    echo "Adding key \$key_name"
    SEED="\$(aws ssm get-parameter --with-decryption --region \$region --name /polkadot/validator-failover/${prefix}/keys/\$key_name/seed | jq -r .Parameter.Value)"
    KEY="\$(aws ssm get-parameter --region \$region --name /polkadot/validator-failover/${prefix}/keys/\$key_name/key | jq -r .Parameter.Value)"
    TYPE="\$(aws ssm get-parameter --region \$region --name /polkadot/validator-failover/${prefix}/keys/\$key_name/type | jq -r .Parameter.Value)"
    
    curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "author_insertKey", "params":["'"\$TYPE"'","'"\$SEED"'","'"\$KEY"'"]}' http://localhost:9933

done
EOF

cat <<EOF >/usr/local/bin/best-grep.sh
#!/bin/bash
BEST=\$(docker logs polkadot 2>&1 | grep finalized | tail -n 1 | cut -d':' -f4 | cut -d'(' -f1 | cut -d'#' -f2 | xargs)
re='^[0-9]+$'
if [[ "\$BEST" =~ \$re ]] ; then
  if [ "\$BEST" -gt 0 ] ; then
    /usr/local/bin/consul kv put best_block "\$BEST"
  else
    echo "Block number either cannot be compared with 0, or not greater than 0"  
  fi 
else
  echo "Block number is not a number, skipping block insertion"
fi

sleep 7

EOF

chmod 700 /usr/local/bin/key-insert.sh
chmod 700 /usr/local/bin/best-grep.sh
chmod 700 /usr/local/bin/double-signing-control.sh

# Create lock for the instance
n=0
set +eE
trap - ERR

until [ $n -ge 6 ]; do

  /usr/local/bin/consul lock prefix "/usr/local/bin/double-signing-control.sh && /usr/local/bin/key-insert.sh && consul kv delete blocks/.lock && consul lock blocks \"while true; do /usr/local/bin/best-grep.sh; done\" & docker stop polkadot && docker rm polkadot && /usr/bin/docker run --cpus $${CPU} --memory $${RAM}GB --kernel-memory $${RAM}GB --name polkadot --restart unless-stopped -p 30333:30333 -p 127.0.0.1:9933:9933 -v /data:/data chevdor/polkadot:latest polkadot --chain ${chain} --unsafe-rpc-external --rpc-cors=all --validator --name '$NAME' --node-key '$NODEKEY'"

  sleep 10;
  n=$[$n+1]

done

set -eE
trap default_trap ERR EXIT

consul leave
shutdown now

# Instance will shutdown when loosing the lock because of no docker container will be running. ASG will replace the instance because it will not pass the ELB health check which monitors all the Consul ports and port 30333 (Polkadot node's port). 
