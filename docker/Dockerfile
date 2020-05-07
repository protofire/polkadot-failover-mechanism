FROM hashicorp/terraform
MAINTAINER 1627_DEV@altoros.com
ADD .. /polkadot-validator-failover
WORKDIR /polkadot-validator-failover
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
