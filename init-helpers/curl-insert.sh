#!/bin/bash 
  
set -x

libs=( libmetalink.so.3 libssl3.so libsmime3.so libnss3.so libnssutil3.so libplds4.so libplc4.so libnspr4.so )
for lib in "${libs[@]}"; do
  docker cp /usr/lib64/${lib} polkadot:/usr/lib/
done

docker cp /usr/bin/curl polkadot:/bin/curl
