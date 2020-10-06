set -o pipefail

IFS=" " read -r -a ports <<<"$1"
rg=$2
vmss=$3
attempts=0
max_attempts=200

command -v az >/dev/null || {
    echo "no az utility...."
    exit 1
}

command -v nc >/dev/null || {
    echo "no nc (netcat) utility...."
    exit 1
}

hosts=()

function get_hosts() {
    cmd="readarray -t hosts < <(az vmss list-instance-public-ips --name ${vmss} --resource-group ${rg} | jq -r '.[].ipAddress')"
    eval "${cmd}" || {
        while true; do
            eval "${cmd}" && break
            sleep 10
        done
    }
}

while true; do
    get_hosts
    if [ "${#hosts}" -eq 0 ]; then
        sleep 10
        continue
    fi
    count=0
    checks_count=$((${#hosts[@]} * ${#ports[@]}))
    for host in "${hosts[@]}"; do
        for port in "${ports[@]}"; do
            echo "checking host ${host} with port ${port}..."
            if ! nc -z -w 1 "${host}" "${port}" &>/dev/null; then
                echo "not available port ${port} on host ${host}"
            else
                echo "host ${host} port is open ${port}..."
                count=$((count + 1))
            fi
        done
    done
    [ ${count} -eq ${checks_count} ] && break
    attempts=$((attempts + 1))
    if [ ${attempts} -eq ${max_attempts} ]; then
        echo "could not check open ports. timeout..."
        exit 1
    else
        echo "Attempts: ${attempts}. Remaining: $((max_attempts - attempts))"
    fi
    sleep 10
done

exit 0
