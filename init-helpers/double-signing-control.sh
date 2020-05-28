#!/bin/bash

set -x

BEST=$(/usr/local/bin/consul kv get best_block)
retVal=$?
set -eE
echo "Previous validator best block - $BEST"

if [ "$retVal" -eq 0 ]; then

  VALIDATED=0
  until [ "$VALIDATED" -gt "$BEST" ]; do

    BEST_TEMP=$(/usr/local/bin/consul kv get best_block)
    if [ "$BEST_TEMP" != "$BEST" ]; then
      consul leave
      shutdown now
      exit 1
    else
      BEST=$BEST_TEMP
      VALIDATED=$(/usr/bin/docker logs polkadot 2>&1 | /usr/bin/grep finalized | /usr/bin/tail -n 1)
      VALIDATED=$(/usr/bin/echo ${VALIDATED##*#} | /usr/bin/cut -d'(' -f1 | /usr/bin/xargs)
      echo "Previous validator best block - $BEST, new validator validated block - $VALIDATED"
      sleep 10
    fi
  done
fi