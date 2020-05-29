#!/usr/bin/env bash

set +e

docker stop polkadot
consul leave
