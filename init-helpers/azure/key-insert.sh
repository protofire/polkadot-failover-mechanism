#!/bin/bash

set -x -e -E

key_names=$(az keyvault secret list --vault-name ${1}-vault | jq .[].name -r | grep '\-keys\-' | cut -d '-' -f4 | uniq)

for key_name in ${key_names[@]} ; do

  echo "Adding key $key_name"
  SEED=$(az keyvault secret show --vault-name ${1}-vault --name polkadot-${1}-keys-${key_name}-seed | jq .value -r)
  KEY=$(az keyvault secret show --vault-name ${1}-vault --name polkadot-${1}-keys-${key_name}-key | jq .value -r)
  TYPE=$(az keyvault secret show --vault-name ${1}-vault --name polkadot-${1}-keys-${key_name}-type| jq .value -r)
   
  docker exec -i polkadot /bin/curl -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_networkState", "params":["'"$TYPE"'","'"$SEED"'","'"$KEY"'"]}' http://localhost:9933

done
