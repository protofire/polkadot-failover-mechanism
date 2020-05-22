# Summary

Docker images were created to simplify the process of using Polkadot Failover Mechanism solution. This drastically simplifies incorporation of that image into CI/CD solutions in particular. Using Docker image operator will not need to install Terraform and other prerequisites, only Docker is required.

# DockerHub

Using automatic CircleCI CD system images getting built with each new release of the Failover Mechanism solution and published at [DockerHub](https://hub.docker.com/repository/docker/protofire/polkadot-failover-mechanism). Anyone can download the image and use it in personal environment.

# Usage example

To set variables required for Terraform launch user can use the Docker built-in capability to pass environment variables inside of the Docker container. Basically, user will need to set all the variables required for Terraform plus `mode` and `iaas` variable.

- `iaas` variable represents the IaaS where solution are going to be deployed. Currently, there are three options supported - `aws`, `azure`, `gcp`. Make sure to set this variables in lowercase.
- `mode` shoud be either `create` or `destroy`, also lowercase.
All the other variables, required (or optional) for each particular IaaS you can find in `terraform.tfvars.example` and `variables.tf` files inside of the IaaS-related folder in this repo. To set them, pass the variable prefixed with `TF_VAR_` prefix, so the `example` variable will be `TF_VAR_example`.

For instance, to launch solution on AWS you can run the following command:
```
docker run protofire/polkadot-failover-mechanism -e iaas=aws -e mode= created -e TF_VAR_prefix=myenv [PUT ALL OTHER REQUIRED VARIABLES HERE]
```