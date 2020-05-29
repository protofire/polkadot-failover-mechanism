# Deployment instruction

## Manual installation

### Prerequisites

1. An instance further referred as *Deployer* instance which will be used to run these scripts from.
2. [Terraform](https://www.terraform.io/downloads.html). To install Terraform proceed to the [Install Terraform](#install-terraform) section.
3. (Optional) GCloud CLI. To install GCloud CLI follow [the instruction](https://cloud.google.com/sdk/docs/quickstarts).

Also you will need a set of keys known as `NODE KEY`, `STASH`, `CONTROLLER` and `SESSION KEYS`. As for this release in `Kusama` and `Westend` network there are 5 keys inside of `SESSION KEYS` object - [GRANDPA, BABE, ImOnline, Parachains, AuthorityDiscovery](https://github.com/paritytech/polkadot/blob/master/runtime/kusama/src/lib.rs#L258). You will have to generate all of them. You can do it either using [Subkey](https://substrate.dev/docs/en/ecosystem/subkey) tool or using [PolkadotJS](https://polkadot.js.org/apps/#/accounts) website.

#### Keys reference

| Key name            | Key short name | Key type |
| ------------------- | -------------- | -------- |
| NODE KEY            | -              | ed25519  |
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

### Create project

It is highly recommended to run these scripts at the dedicated project. Despite the fact that security rules are configured to limit deployed service accounts as much as possible there are some CLI requests like secrets listing that can't be constrained. For better security and logical distribution.

### Clone the repo

Either clone this repo using `git clone` command or simply download it from Web and unpack on the deployer node.

### Run the Terraform scripts

1. Open `gcp` folder of the cloned (downloaded) repo.
2. Create `terraform.tfvars` file inside of the `gcp` folder of the cloned repo, where `terraform.tfvars.example` is located.
3. Fill it with the appropriate variables. You can check the very minimum example at [example](terraform.tfvars.example) file and the full list of supported variables (and their types) at [variables](variables.tf) file. Fill `validator_keys` variable with your SESSION KEYS. For key types use short types from the following table - [Keys reference](#keys-reference).
5. Run `terraform init`.
6. Run `terraform plan -out terraform.tfplan` and check the set of resources to be created on your cloud account.
7. If you are okay with the proposed plan - run `terraform apply terraform.tfplan` to apply the deployment.
8. After the deployment is complete you can open Azure Portal to check that the instances were deployed successfully.

*!IMPORTANT!* Unlike AWS and Azure GCP will not provide you with alarm each time failover occurs due to metrics sending methods currently implemented. #5 issue should fix this.

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

# Known issues & limitations

## Prefix should contain alphanumeric characters only and have to be short

The prefix is used in a majority of resources names, so they can be easily identified among others. This causes the limitation because not all of the deployed resources supports long names or names with non alphanumeric symbols. The optimal is to have around 5 alphanumeric characters as a system prefix.

## Multi-regional provider outage

As for now the implemented failover mechanism won't work if 2 out of the 3 chosen regions goes offline. Make sure to use geographically distributed regions to improve nodes stability.

## Not all disks are deleted after infrastructure is deleted

Set `delete_on_terminate` variable to `true` to override this behavior.
