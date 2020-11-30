#!/bin/sh

set -e

: "${mode:=create}"
: "${iaas:=aws}"

if [ -z "${mode}" ]; then
  echo No mode specified
  exit 1
fi

if [ -z "${iaas}" ]; then
  echo No IAAS specified
  exit 1
fi

if ! [ -d "${iaas}" ]; then
  echo Incorrect IAAS specified
  exit 1
fi

if [ ${mode} != "create" ] && [ ${mode} != "delete" ]; then
  echo Unknown mode: "${mode}"
  exit 1
fi

cd "${iaas}" || exit 1

terraform init

if [ ${mode} = "create" ]; then
  terraform plan -out terraform.tfplan
  terraform apply -auto-approve terraform.tfplan
  exit 0
fi

if [ ${mode} = "destroy" ]; then
  terraform destroy -auto-approve
  exit 0
fi
