#!/bin/sh

if [ -z ${mode} ]; then
  echo No mode specified
  exit 1
fi

if [ -z ${iaas} ]; then
  echo No IAAS specified
  exit 1
fi

if ! [ -d ${iaas} ]; then
  echo Incorrect IAAS specified
  exit 1
fi

cd ${iaas}
terraform init
if [[ ${mode} == "create" ]]; then
  terraform plan -out terraform.tfplan
  terraform apply -auto-approve terraform.tfplan
  exit 0
fi

if [[ ${mode} == "destroy" ]]; then
  terraform destroy -auto-approve
  exit 0
fi
echo Unknown mode: ${mode}
