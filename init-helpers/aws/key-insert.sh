#!/bin/bash

set -x -e -E

region="$(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region)"
key_names=$(aws ssm get-parameters-by-path --region $region --recursive --path /polkadot/validator-failover/${1}/keys/ | jq .Parameters[].Name | awk -F'/' '{print$(NF-1)}' | sort | uniq)

for key_name in ${key_names[@]} ; do

    echo "Adding key $key_name"
    SEED="$(aws ssm get-parameter --with-decryption --region $region --name /polkadot/validator-failover/${1}/keys/${key_name}/seed | jq -r .Parameter.Value)"
    KEY="$(aws ssm get-parameter --region $region --name /polkadot/validator-failover/${1}/keys/${key_name}/key | jq -r .Parameter.Value)"
    TYPE="$(aws ssm get-parameter --region $region --name /polkadot/validator-failover/${1}/keys/${key_name}/type | jq -r .Parameter.Value)"
    
    docker exec -i polkadot /bin/curl -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_networkState", "params":["'"$TYPE"'","'"$SEED"'","'"$KEY"'"]}' http://localhost:9933

done
