#!/bin/bash
set -x -e -E

# Fetch keys from SSM. Then add the fetched keys to the node via the curl request
readarray -t key_names < <(gcloud secrets list --format=json --filter="name ~ ${1}_ AND labels.prefix=${1} AND labels.type=key" | jq .[].name -r | cut -d '_' -f2 | uniq)

for key_name in "${key_names[@]}"; do

  echo "Adding key '$key_name'"

  SEED_NAME=$(gcloud secrets list --format=json --filter="name ~ ${1}_${key_name}_seed AND labels.prefix=${1} AND labels.type=key" | jq -r .[0].name)
  KEY_NAME=$(gcloud secrets list --format=json --filter="name ~ ${1}_${key_name}_key AND labels.prefix=${1} AND labels.type=key" | jq -r .[0].name)
  TYPE_NAME=$(gcloud secrets list --format=json --filter="name ~ ${1}_${key_name}_type AND labels.prefix=${1} AND labels.type=key" | jq -r .[0].name)

  ACCESS_SEED_NAME=$(gcloud secrets versions list "${SEED_NAME}" --format json | jq '.[] | select(.state == "ENABLED") | .name' -r)
  ACCESS_KEY_NAME=$(gcloud secrets versions list "${KEY_NAME}" --format json | jq '.[] | select(.state == "ENABLED") | .name' -r)
  ACCESS_TYPE_NAME=$(gcloud secrets versions list "${TYPE_NAME}" --format json | jq '.[] | select(.state == "ENABLED") | .name' -r)

  SEED=$(gcloud secrets versions access "${ACCESS_SEED_NAME}" --secret="${SEED_NAME}" --format json | jq .payload.data -r | base64 -d)
  KEY=$(gcloud secrets versions access "${ACCESS_KEY_NAME}" --secret="${KEY_NAME}" --format json | jq .payload.data -r | base64 -d)
  TYPE=$(gcloud secrets versions access "${ACCESS_TYPE_NAME}" --secret="${TYPE_NAME}" --format json | jq .payload.data -r | base64 -d)
  
  curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "author_insertKey", "params":["'"$TYPE"'","'"$SEED"'","'"$KEY"'"]}' http://localhost:9933
done
