[ $# -lt 4 ] && {
    echo "Subscription ID, resource group, scale set name and metric params required as script arguments"
    exit 1
}

set -e -o pipefail

flag=false

function echoErr() { printf "%s\n" "$*" >&2; }

function stop() {
    echoErr "Stopping script..."
    flag=true
}

trap stop TERM INT

RESOURCE="/subscriptions/$1/resourceGroups/$2/providers/Microsoft.Compute/virtualMachineScaleSets/$3"
shift 3

max_attempts=300
attempts=0
recheck_metrics_count=10
success_metric_checks=0

declare -A metrics=()

function get_metric_name_filter() {
    local name=${1}
    local title_name="${1^}"
    local upper_name="${1^^}"
    echo "[].name.value | ([?@=='${name}'] || [?@=='${title_name}'] || [?@=='${upper_name}']) | [0]"
}

function check_metric() {
    local param=${1}
    local metric=${2}
    local query
    query=$(get_metric_name_filter "${metric}")
    value=$(az monitor metrics list-definitions --resource "${RESOURCE}" --namespace "${namespace}" --query "${query}" | grep -i "${metric}")
    # shellcheck disable=SC2181
    if [ $? -ne 0 ]; then
        echo ""
    else
        # shellcheck disable=SC2001
        echo "${value}" | sed 's/"//g'
    fi

}

function check_namespaces() {
    local params=("${@}")
    local count=0
    for param in "${params[@]}"; do
        local parts
        IFS=':' read -r -a parts <<<"${param}"
        local namespace="${parts[0]}"
        local metric="${parts[1]}"
        local real_metric_name
        real_metric_name=$(check_metric "${namespace}" "${metric}")
        if [ -z "${real_metric_name}" ]; then
            echoErr "Cannot find metric ${metric} in namespace ${namespace}"
        else
            echoErr "Found metric ${metric} in namespace ${namespace}"
            metrics[${namespace}]=${real_metric_name}
            count=$((count + 1))
        fi
    done
    [ ${count} -eq ${#params[@]} ] && {
        success_metric_checks=$((success_metric_checks + 1))
        if [ ${success_metric_checks} -eq ${recheck_metrics_count} ]; then
            echoErr "Ensured all namespace metrics"
            return 0
        fi
        echoErr "Found all namespace metrics. Rechecking ${success_metric_checks} from ${recheck_metrics_count}..."
        return 1
    }
    success_metric_checks=0
    return 1
}

function assoc2json() {
    printf '%s\0' "${!metrics[@]}" "${metrics[@]}" |
        jq -Rs 'split("\u0000") | . as $v | (length / 2) as $n | reduce range($n) as $idx ({}; .[$v[$idx]]=$v[$idx+$n])'
}

while ! check_namespaces "${@}"; do
    [ "${flag}" == "true" ] && exit 1
    attempts=$((attempts + 1))
    [ ${attempts} -eq ${max_attempts} ] && {
        echoErr "Cannot find all namespace metrics"
        exit 1
    }
    sleep 10
done

assoc2json

exit 0
