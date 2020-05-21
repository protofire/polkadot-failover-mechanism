# Deployment instruction

## Manual installation

### Prerequisites

You will need an instance supported by [Terraform](https://www.terraform.io/downloads.html) further referred as *Deployer* instance. No specific requirements for these particular scripts needed.

Also you will need a set of keys known as `STASH`, `CONTROLLER` and `SESSION KEYS`. As for this release in `Kusama` and `Westend` network there are 5 keys inside of `SESSION KEYS` object - [GRANDPA, BABE, ImOnline, Parachains, AuthorityDiscovery](https://github.com/paritytech/polkadot/blob/master/runtime/kusama/src/lib.rs#L258). You will have to generate all of them. You can do it either using [Subkey](https://substrate.dev/docs/en/ecosystem/subkey) tool or using [PolkadotJS](https://polkadot.js.org/apps/#/accounts) website.

#### Keys reference

| Key name            | Key short name | Key type |
| ------------------- | -------------- | -------- |
| STASH               | -              | sr25519  |
| CONTROLLER          | -              | ed25519  |
| GRANDPA             | gran           | ed25519  |
| BABE                | babe           | sr25519  |
| I'M ONLINE          | imon           | sr25519  |
| PARACHAINS          | para           | sr25519  |
| AUTHORITY DISCOVERY | audi           | sr25519  |

### Install Terraform

1. Download [Terraform](https://www.terraform.io/downloads.html).
2. Unpack Terraform using `unzip terraform*` command.
3. Move the `terraform` binary to the one of the folders specified at the `PATH` variable. For example: `sudo mv terraform /usr/local/bin/`

### Clone the repo

Either clone this repo using `git clone` command or simply download it from Web and unpack on the deployer node.

### Run the Terraform scripts

1. Open `aws` folder of the cloned (downloaded) repo.
2. Create `terraform.tfvars` file inside of the `aws` folder of the cloned repo, where `terraform.tfvars.example` is located
3. Fill it with the appropriate variables. You can check the very minimum example at [example](terraform.tfvars.example) file and the full list of supported variables (and their types) at [variables](variables.tf) file. Fill `validator_keys` variable with your SESSION KEYS. For key types use short types from the following table - [Keys reference](#keys-reference).
4. Set `AWS_ACCESS_KEY` and `AWS_SECRET_KEY` environment variables.
5. (Optional) You can either place a Terraform state file on S3 bucket or on your local machine. To place it on the local machine rename the `remote-state.tf` file to `remote-state.tf.stop`. To place it on S3 - create an S3 bucket and proceed to the next step. You will be interactively asked to provide S3 configuration details.
6. Run `terraform init`
7. Run `terraform plan -out terraform.tfplan` and check the set of resources to be created on your cloud account.
8. If you are okay with the proposed plan - run `terraform apply terraform.tfplan` to apply the deployment.
9. After the deployment is complete you can open your EC2 console to check that the instances were deployed successfully. You can also open Polkadot explorer and ensure that your node is in the validators list.

### Validate

1. Watch [Polkadot Telemetry](https://telemetry.polkadot.io/) for your node to synchronize with the network.<br />
2. Make sure you have funds on your STASH account. Bond your fund to CONTROLLER account. For this and the following steps you can either perform a transaction on your node or use or use [PolkadotJS](https://polkadot.js.org/apps/#/staking/actions) website. For this operation use `staking.bond` transaction.
3. Set your session keys to the network - perform a `session.setKeys` transaction. As an argument pass all your session keys in hex format in a order specified [here](https://github.com/paritytech/polkadot/blob/master/runtime/kusama/src/lib.rs#L258) concatenating them one by one. 
```
For example if you have the following keys:
GRAN - 0xbeaa0ec217371a8559f0d1acfcc4705b48082b7a02fd6cb2e76714380576151e
BABE - 0xec648f4ad1693cc61e340aa122c7142d7603e26e04a47a5f0811c31a60c07b49
IMON - 0x9633780f889f0fc6280adba40695139f77c00e53168544492c6fa2399b693e3c
PARA - 0xee383120ff7b87409e105de2b0150432a95153d0a1edd5bea0af669001b80f1d
AUDI - 0x701ed6b86f109a6d59d7933df3311c5b6edc3862657179259cb983149bfc404c

The argument for sessions.setKeys will be 0xbeaa0ec217371a8559f0d1acfcc4705b48082b7a02fd6cb2e76714380576151eec648f4ad1693cc61e340aa122c7142d7603e26e04a47a5f0811c31a60c07b499633780f889f0fc6280adba40695139f77c00e53168544492c6fa2399b693e3cee383120ff7b87409e105de2b0150432a95153d0a1edd5bea0af669001b80f1d701ed6b86f109a6d59d7933df3311c5b6edc3862657179259cb983149bfc404c

Note that there is only one 0x left, all the others are omitted.
```
4. Start validating - perform a `staking.validate` transaction.
5. Subscribe to notifications at AWS Simple Notifications Service to start receiving alarms from your nodes.

# Known issues & limitations

## Prefix should contain alphanumeric characters and have to be short

The prefix is used in a majority of resources names, so they can be easily identified among others. This causes the limitation because not all of the deployed resources supports long names or names with non alphanumeric symbols. The optimal is to have around 5 alphanumeric characters as a system prefix.

## Node not getting started on t2 (and some other) instances

Current version of failover scripts supports only the [Nitro instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html#ec2-nitro-instances) as the node type. This is connected to the way disks are being managed by AWS - nitro instances has the most disks attached as the `/dev/xvd*` which does not have the very same numeration as on the EC2 console. This issue has to be fixed in future, but we are open for any improvements PRs for now.

## Multi-regional provider outage

As for now the implemented failover mechanism won't work if 2 out of the 3 chosen regions goes offline. Make sure to use geographically distributed regions to improve nodes stability.

## Not all disks are deleted after infrastructure is deleted

This is the intended behavior. Architecture were designed to minimize the risk of data loss, so the validator would not need to resync blockchain data. The disks that sustained the infrastructure destruction contains Polkadot data. Thus, the very same disks can be used when re-raising the infrastructure. If data discs are no longer needed, they can be deleted manually. To do that simply filter them using the `prefix` tag. All the infrastructure components are tagged by `prefix` tag during the deployment.

You can also override this default behavior by setting `delete_on_terminate` variable to `true`.

# Proposed improvements

## Spot instances

The average failure rate for Spot instances is around 1-15% hourly. It can be decreased using the AWS Launch Template with an array of supported Spot instance types. The failover mechanism will reelect another node as the cluster leader, so there will be just a couple of seconds of downtime. Thus, we will drastically reduce the infrastructure costs with a minor decrease of failoverability provided by these scripts.
