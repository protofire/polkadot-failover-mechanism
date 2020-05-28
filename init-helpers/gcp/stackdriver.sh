#!/usr/bin/env bash
set -ex

function generate_collectd_checks {

## Stackdriver configs

SCRIPTS_PATH=/opt/stackdriver/collectd/custom/
mkdir -p $SCRIPTS_PATH 

# Metric generator scripts, one for each metric

# status
cat <<EOF > $SCRIPTS_PATH/status.sh
#! /bin/bash
INTERVAL="\$COLLECTD_INTERVAL"
HOSTNAME="\$COLLECTD_HOSTNAME"

while  true; do
  docker exec -i polkadot /bin/curl --connect-timeout 2 -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_health", "params":[]}' http://127.0.0.1:9933;
  STATE=$?

  echo "PUTVAL \${HOSTNAME}/exec-polkadot/gauge-state/ interval=\${INTERVAL} N:\${STATE}"
  sleep \${INTERVAL}

done

EOF

# blocknumber
cat <<EOF > $SCRIPTS_PATH/blocknumber.sh
#! /bin/bash
INTERVAL="\$COLLECTD_INTERVAL"
HOSTNAME="\$COLLECTD_HOSTNAME"

while true; do

  BLOCK_NUMBER_HEX=\$(docker exec -i polkadot /bin/curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "chain_getBlock", "params":[]}' http://127.0.0.1:9933 | jq .result.block.header.number -r)


  BLOCK_NUMBER=\$(( 16#$(echo \$BLOCK_NUMBER_HEX | sed 's/^0x//')))

  echo "PUTVAL \${HOSTNAME}/exec-polkadot/gauge-blocknumber/ interval=\${INTERVAL} N:\${BLOCK_NUMBER}"

  sleep \${INTERVAL}
done

EOF

# validatorcount
cat <<EOF > $SCRIPTS_PATH/validatorcount.sh
#! /bin/bash
INTERVAL="\$COLLECTD_INTERVAL"
HOSTNAME="\$COLLECTD_HOSTNAME"

while true ; do

  AMIVALIDATOR=\$(docker exec -i polkadot /bin/curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_nodeRoles", "params":[]}' http://127.0.0.1:9933 | jq -r .result[0])
  if [ "\$AMIVALIDATOR" == "Authority" ]; then
    AMIVALIDATOR=1
  else
    AMIVALIDATOR=0
  fi

  echo "PUTVAL \${HOSTNAME}/exec-polkadot/gauge-validatorcount/ interval=\${INTERVAL} N:\${AMIVALIDATOR}"

  sleep \${INTERVAL}
done
EOF

pushd $SCRIPTS_PATH
  chown nobody blocknumber.sh status.sh validatorcount.sh
  chmod 700 blocknumber.sh status.sh validatorcount.sh
popd

}


function  generate_collectd_rewrite {

# CollectD rewrite URLs

cat <<EOF > /etc/stackdriver/collectd.d/test.conf

LoadPlugin exec

<Plugin "exec">
    Exec "nobody" "$SCRIPTS_PATH/status.sh"
</Plugin>

<Plugin "exec">
    Exec "nobody" "$SCRIPTS_PATH/validatorcount.sh"
</Plugin>

<Plugin "exec">
    exec "nobody" "$SCRIPTS_PATH/blocknumber.sh"
</Plugin>

LoadPlugin target_set
PreCacheChain "PreCache"
<Chain "PreCache">
  <Rule "polkadot">
    <Match regex>
      Plugin "^exec$"
      PluginInstance "^polkadot$"
    </Match>
    <Target "set">
      MetaData "stackdriver_metric_type" "custom.googleapis.com/polkadot/%{plugin_instance}/%{type_instance}"
      MetaData "label:name"  "polkadot"
      MetaData "label:hostname" "%{host}"
    </Target>
  </Rule>
</Chain>

EOF

}
