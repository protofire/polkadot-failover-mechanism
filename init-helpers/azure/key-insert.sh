#!/bin/bash

set -x -e -E
set -o pipefail

KEY_VAULT_NAME="${1}"
PREFIX="${2}"

readarray -t key_names < <(az keyvault secret list --vault-name "${KEY_VAULT_NAME}" | jq '.[].name' -r | grep '\-keys\-' | cut -d '-' -f4 | uniq)

for key_name in "${key_names[@]}"; do

    echo "Adding key $key_name"

    SEED=$(az keyvault secret show --vault-name "${KEY_VAULT_NAME}" --name "polkadot-${PREFIX}-keys-${key_name}-seed" | jq .value -r)
    KEY=$(az keyvault secret show --vault-name "${KEY_VAULT_NAME}" --name "polkadot-${PREFIX}-keys-${key_name}-key" | jq .value -r)
    TYPE=$(az keyvault secret show --vault-name "${KEY_VAULT_NAME}" --name "polkadot-${PREFIX}-keys-${key_name}-type" | jq .value -r)

    curl -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "author_insertKey", "params":["'"$TYPE"'","'"$SEED"'","'"$KEY"'"]}' http://localhost:9933

done
