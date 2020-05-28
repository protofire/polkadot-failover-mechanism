#!/bin/bash

$(docker inspect -f "{{.State.Running}}" polkadot && docker exec -i polkadot /bin/curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_health", "params":[]}' http://127.0.0.1:9933);
STATE=$?
BLOCK_NUMBER=$(($(docker exec -i polkadot /bin/curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "chain_getBlock", "params":[]}' http://127.0.0.1:9933 | jq .result.block.header.number -r)))
AMIVALIDATOR=$(docker exec -i polkadot /bin/curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_nodeRoles", "params":[]}' http://127.0.0.1:9933 | jq -r .result[0])
if [ "$AMIVALIDATOR" == "Authority" ]; then
  AMIVALIDATOR=1
else
  AMIVALIDATOR=0
fi

regions=( $3 $4 $5 )

for i in "${regions[@]}"; do
  aws cloudwatch put-metric-data --region $i --metric-name "Health report" --dimensions AutoScalingGroupName=$2 --namespace "${1}" --value "$STATE"
  aws cloudwatch put-metric-data --region $i --metric-name "Block Number" --dimensions InstanceID="$(curl --silent http://169.254.169.254/latest/meta-data/instance-id)" --namespace "${1}" --value "$BLOCK_NUMBER"
  aws cloudwatch put-metric-data --region $i --metric-name "Validator count" --dimensions AutoScalingGroupName=$2 --namespace "${1}" --value "$AMIVALIDATOR"
done
