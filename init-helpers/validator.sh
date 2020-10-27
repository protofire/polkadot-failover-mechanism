#!/usr/bin/env bash

function start_polkadot_passive_mode () {

  if [ "$#" -lt 6 ]; then
    echo "argument $*, number $#. Should not be less then 6"
    return 1
  fi

  local docker_name="$1"
  local cpu="$2"
  local ram="$3"
  local image="$4"
  local chain="$5"
  local data="$6"
  local unsafe="${7:-false}"

  local cmd=( "${image}" "--chain=${chain}" "--pruning=archive" "-d" "${data}" )

  if [ "${unsafe}" = true ]; then
    cmd+=( "--rpc-methods=Unsafe" "--rpc-external" )
  fi

  /usr/bin/docker stop "${docker_name}" || true
  /usr/bin/docker rm -f "${docker_name}" || true

  /usr/bin/docker run \
  --cpus "${cpu}" \
  --memory "${ram}" \
  --kernel-memory "${ram}" \
  --network=host \
  --name "${docker_name}" \
  --restart unless-stopped \
  -d \
  -p 127.0.0.1:9933:9933 \
  -p 30333:30333 \
  -v "${data}:${data}:z" \
  "${cmd[@]}"
}

function start_polkadot_validator_mode () {

  if [ "$#" -lt 8 ]; then
    echo "argument $*, number $#. Should not be less then 8"
    return 1
  fi

  local docker_name="$1"
  local cpu="$2"
  local ram="$3"
  local image="$4"
  local chain="$5"
  local data="$6"
  local validator_name="$7"
  local validator_node_key="$8"

  /usr/bin/docker stop "${docker_name}"
  /usr/bin/docker rm -f "${docker_name}"

  /usr/bin/docker run \
  --cpus "${cpu}" \
  --memory "${ram}" \
  --kernel-memory "${ram}" \
  --network=host \
  --name "${docker_name}" \
  --restart unless-stopped \
  -p 127.0.0.1:9933:9933 \
  -p 30333:30333 \
  -v "${data}:${data}:z" \
  "${image}" --chain "${chain}" --validator --name "${validator_name}" --node-key "${validator_node_key}" -d "${data}"
}
