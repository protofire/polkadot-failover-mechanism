#!/bin/bash

BEST=$(docker logs polkadot 2>&1 | grep finalized | tail -n 1 | cut -d':' -f4 | cut -d'(' -f1 | cut -d'#' -f2 | xargs)
re='^[0-9]+$'
if [[ "$BEST" =~ $re ]] ; then
  if [ "$BEST" -gt 0 ] ; then
    /usr/local/bin/consul kv put best_block "$BEST"
  else
    echo "Block number either cannot be compared with 0, or not greater than 0"  
  fi 
else
  echo "Block number is not a number, skipping block insertion"
fi

sleep 7