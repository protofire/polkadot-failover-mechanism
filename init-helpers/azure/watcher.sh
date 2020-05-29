#!/bin/bash

$(docker inspect -f "{{.State.Running}}" polkadot && curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_health", "params":[]}' http://127.0.0.1:9933);
STATE=$?
BLOCK_NUMBER=$(($(curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "chain_getBlock", "params":[]}' http://127.0.0.1:9933 | jq .result.block.header.number -r)))
AMIVALIDATOR=$(curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_nodeRoles", "params":[]}' http://127.0.0.1:9933 | jq -r .result[0])
if [ "$AMIVALIDATOR" == "Authority" ]; then
  AMIVALIDATOR=1
else
  AMIVALIDATOR=0
fi

ts=$(timestamp)

curl -i -XPOST 'http://localhost:12500/telegraf' --data-binary "health value=$STATE"
curl -i -XPOST 'http://localhost:12500/telegraf' --data-binary "block value=$BLOCK_NUMBER"
curl -i -XPOST 'http://localhost:12500/telegraf' --data-binary "validator value=$AMIVALIDATOR"
