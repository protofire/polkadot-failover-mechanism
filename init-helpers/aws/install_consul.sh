# Helper function to get IP address of consul instance from metadata server

# Clone and install consul

function join_by { local IFS="$1"; shift; echo "$*"; }

function install_consul() {

  git clone https://github.com/hashicorp/terraform-aws-consul.git
  terraform-aws-consul/modules/install-consul/install-consul --version 1.7.2

  set -ex
  # set up parameters to configure consul installation
  echo "Consul version: ${CONSUL_VERSION:=1.7.2}"
  local -r CONSUL_USER=consul
  local -r CONSUL_DATA_DIR="/var/lib/consul"
  local -r CONSUL_CONFIG_DIR="/opt/consul/config"
  local -r CONSUL_CONFIG_FILE="new-retry-join.json"
  local -r CONFIG_PATH="$CONSUL_CONFIG_DIR/$CONSUL_CONFIG_FILE"

  prefix="$1"
  total_instance_count=$2
  lb_primary="$3"
  lb_secondary="$4"
  lb_tertiary="$5"

  # Add consul user
  id $CONSUL_USER >& /dev/null || useradd --system $CONSUL_USER

  # Create and chown folders for consul
  mkdir -p $CONSUL_DATA_DIR
  chown -R $CONSUL_USER:$CONSUL_USER $CONSUL_DATA_DIR
  mkdir -p $CONSUL_CONFIG_DIR
  chown -R $CONSUL_USER:$CONSUL_USER $CONSUL_CONFIG_DIR

  lbs=()

  [ -n "${lb_primary}" ] && lbs+=( "\"${lb_primary}\"" )
  [ -n "${lb_secondary}" ] && lbs+=( "\"${lb_secondary}\"" )
  [ -n "${lb_tertiary}" ] && lbs+=( "\"${lb_tertiary}\"" )

  lbs_str=$(join_by ", ", "${lbs[@]}")

  echo "LoadBalancers: ${lbs_str}"
  ## Override consul config. Add join through load balancers and leave_on_terminate flag
  cat <<EOF > "$CONFIG_PATH"
{
  "retry_join": [
    $lbs_str
  ],
  "leave_on_terminate": true,
  "bootstrap_expect": ${total_instance_count}
}
EOF

  # Start consul with a new script in server mode
  /opt/consul/bin/run-consul --server --datacenter "${prefix}"

}

# Function to generate consul health checks to validate node health
function create_consul_healthcheck {

  cat <<EOF > "${CONSUL_CONFIG_DIR}/healthcheck.json"
{
   "check": {
      "id": "consul_check",
      "name": "Check Consul health 9500",
      "tcp": "localhost:9500",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "8600",
      "name": "Check health 8600",
      "tcp": "localhost:8600",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "8300",
      "name": "Check health 8300",
      "tcp": "localhost:8300",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "8301",
      "name": "Check health 8301",
      "tcp": "localhost:8301",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "8302",
      "name": "Check health 8302",
      "tcp": "localhost:8302",
      "interval": "10s",
      "timeout": "1s"
    },
    "check": {
      "id": "30333",
      "name": "Check health 30333",
      "tcp": "localhost:30333",
      "interval": "10s",
      "timeout": "1s"
    }
}
EOF

chown "${CONSUL_USER}:${CONSUL_USER}" "${CONSUL_CONFIG_DIR}/healthcheck.json"

}

create_consul_healthcheck