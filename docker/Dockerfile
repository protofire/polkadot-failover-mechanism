FROM golang:alpine as builder
ENV USER=terraform APPNAME=polkadot USER_ID=1000 APPDIR=polkadot-validator-failover CGO_ENABLED=0
ADD . /${APPDIR}
RUN apk add make git curl && adduser -D -h /${APPDIR} -u ${USER_ID} ${USER} && cd /${APPDIR} && su ${USER} -c "make install-all"

FROM hashicorp/terraform:0.13.5
MAINTAINER 1627_DEV@altoros.com
ENV USER=terraform APPNAME=polkadot APPDIR=polkadot-validator-failover
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /${APPDIR}/.terraform.d /$APPDIR
ADD . /${APPDIR}
RUN chown -R ${USER} /${APPDIR}
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh
WORKDIR /${APPDIR}
USER ${USER}
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
