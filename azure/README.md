# Beta-release note
According to [Azure docs](https://github.com/MicrosoftDocs/azure-docs/blob/master/articles/virtual-machines/linux/using-cloud-init.md) the CentOS image with cloud-init scripts that is used in this solution is still in beta. Despite the fact that these scripts were tested on Azure the new versions of image might introduce breaking changes that will affect the scripts behavior.

# Deployment instruction

## Manual installation

### Prerequisites

1. An instance further referred as *Deployer* instance which will be used to run these scripts from.
2. [Terraform](https://www.terraform.io/downloads.html). To install Terraform proceed to the [Install Terraform](#install-terraform) section.
3. Azure CLI. To install Azure CLI follow [the instruction](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest).

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
****
### Create Azure resource group

It is highly recommended to run these scripts at the dedicated resource group. 
1. Login to [Azure Portal](https://portal.azure.com)
2. Type "Resource groups" in the search bar to navigate to the Resource Group management section
3. Click "Add" to create a new Resource group. Note down the name of the created resource group as you will need it further.

### Install Azure az utility

#### YUM based

    sudo rpm --import https://packages.microsoft.com/keys/microsoft.asc
    sudo sh -c 'echo -e "[azure-cli]
    name=Azure CLI
    baseurl=https://packages.microsoft.com/yumrepos/azure-cli
    enabled=1
    gpgcheck=1
    gpgkey=https://packages.microsoft.com/keys/microsoft.asc" > /etc/yum.repos.d/azure-cli.repo'
    sudo yum install -y azure-cli
    
#### Ubuntu/Debian

    curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

#### Login

    az login

### Get Azure credentials

To get the [service account credentials](https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal) that will be used to run these scripts you will need to create an application in Azure Active Directory.
1. Login to [Azure Portal](https://portal.azure.com)
2. Type "Subscriptions" in the search bar to navigate to the Subscriptions management section. Remember the ID of the subscription you want to create resources in. You can read more about Azure entities [here](https://techcommunity.microsoft.com/t5/azure/understanding-azure-account-subscription-and-directory/td-p/34800).
3. Type "Azure Active Directory" (AAD) in the search bar. Navigate to "App registrations" tab inside of the AAD management section.
4. Create new application. Type any recognizable name you want.
5. Save tenant id and client id after creation in your notes. You will need it later.
6. Go to the "Certificates & Secrets" tab inside of the application management pane. Create new client secret and save it. You will need this later as well.

### Get Azure storage account

    az storage account create -n account -g MyResourceGroup
    az storage account keys list -n account -g MyResourceGroup

### Clone the repo

You will need to clone this repo and it's submodules. Use `git clone --recurse-submodules` command, `--recurse-submodules` is mandatory to make scripts work.

### Run the Terraform scripts

1. Login to Azure CLI. Run `az login` command and follow the instructions on the screen.
2. Open `azure` folder of the cloned (downloaded) repo.
3. Create `terraform.tfvars` file inside of the `azure` folder of the cloned repo, where `terraform.tfvars.example` is located.
4. Fill it with the appropriate variables. You can check the very minimum example at [example](terraform.tfvars.example) file and the full list of supported variables (and their types) at [variables](variables.tf) file. Fill `validator_keys` variable with your SESSION KEYS. For key types use short types from the following table - [Keys reference](#keys-reference).
5. Run `terraform init`.
In case you use azure service principal, copy backend/azurerm.tf.example into backend/azurerm.tf,
fill variables and run `terraform init -backend-config=backend/azurerm.tf`
6. Run `terraform plan -out terraform.tfplan` and check the set of resources to be created on your cloud account.
7. If you are okay with the proposed plan - run `terraform apply terraform.tfplan` to apply the deployment.
8. After the deployment is complete you can open Azure Portal to check that the instances were deployed successfully.

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

# Operations

## What alerts will I receive on my email?

You will receive the alerts in the following cases:
- Node reports unhealthy status

*Important!* Since there is no existing way to aggregate multiple VMSS to single metric there is no way to create an alert that will ensure there is only 1 active validator at a time. If you have an idea of how to create this alert - please, open issue or pull request against this repo. Thanks!

## How can I know which node is taking the lead right now?

Basically, there are two possible ways to understand which node is taking the lead. First is to go to the [Azure Monitor dashboard](https://portal.azure.com/#blade/Microsoft_Azure_Monitoring/AzureMonitoringBrowseBlade/overview), and check each of the VMSS. It will have a custom "polkadot/validator" metric, and only one VMSS will report that it runs a validator. In case you are running more than 3 nodes you will not be able to identify validator using this method.

The other way is to SSH into each node subsequentally and run `sudo docker ps -a --no-trunc` command. This command will show you the docker container that are run on this machine. Check the command that is used to run the container. Only one container on one instance will have `--validator` argument at the launch command. All the other containers will have the `--pruning=archive`.

# Known issues & limitations

## Prefix should contain alphanumeric characters only and have to be short

The prefix is used in a majority of resources names, so they can be easily identified among others. This causes the limitation because not all of the deployed resources supports long names or names with non alphanumeric symbols. The optimal is to have around 5 alphanumeric characters as a system prefix.

## Multi-regional provider outage

As for now the implemented failover mechanism won't work if 2 out of the 3 chosen regions goes offline. Make sure to use geographically distributed regions to improve nodes stability.

## Not all disks are deleted after infrastructure is deleted

Set `delete_on_terminate` variable to `true` to override this behavior.

## Nodes are not getting started with "Failed to listen on MultiAddress" error

We faced this issue when set the abnormally long validator name with a number of spaces in it. While the roots of that issue remains unclear we recommend to set validator name using alphanumeric characters only with the length of around 10 symbols.
