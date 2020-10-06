[ $# -lt 4 ] && {
    echo "Subscription ID, resource group, scale set name, prefix and metric params required as script arguments"
    exit 1
}

set -e -o pipefail

flag=false

function stop {
    echo "Stopping script..."
    flag=true
}


trap stop TERM INT


RESOURCE="/subscriptions/$1/resourceGroups/$2/providers/Microsoft.Compute/virtualMachineScaleSets/$3"
shift 3

max_attempts=500
attempts=0

function echoErr() { printf "%s\n" "$*" >&2; }

function get_metric_name_filter() {
    local name=${1}
    local lower_name="${1,,}"
    echo "[].name.value | ([?@=='${name}'] || [?@=='${lower_name}']) | [0]"
}

function check_namespace() {
    local param=${1}
    local metric=${2}
    local query
    query=$(get_metric_name_filter "${metric}")
    az monitor metrics list-definitions --resource "${RESOURCE}" --namespace "${namespace}" --query "${query}" | grep -i "${metric}" &>/dev/null && return 0
    return 1
}

function check_namespaces() {
    local params=("${@}")
    local count=0
    for param in "${params[@]}"; do
        local parts
        IFS=':' read -r -a parts <<<"${param}"
        local namespace="${parts[0]}"
        local metric="${parts[1]}"
        if ! check_namespace "${namespace}" "${metric}"; then
            echoErr "Cannot find metric ${metric} in namespace ${namespace}"
        else
            echo "Found metric ${metric} in namespace ${namespace}"
            count=$((count + 1))
        fi
    done
    [ ${count} -eq ${#params[@]} ] && {
        echo "Found all namespace metrics"
        return 0
    }
    return 1
}

while ! check_namespaces "${@}"; do
    [ "${flag}" == "true" ] && exit 1
    attempts=$((attempts + 1))
    [ ${attempts} -eq ${max_attempts} ] && {
        echo "Cannot find all namespace metrics"
        exit 1
    }
    sleep 10
done

exit 0
