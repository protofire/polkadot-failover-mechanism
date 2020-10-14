polkadot tests
=================

#### Prerequisites

* Install **terraform** >= 13.4
* Install **go** >= 1.15.2
* Install **jq** utility 

### By provider

#### AWS

Use 

    aws configure
    
or set next environment variables:

`AWS_ACCESS_KEY`

`AWS_SECRET_KEY`

##### Command:
    
    make aws
    
#### GCP

Use 

    gcloud auth login

or set next environment variables:

`GCP_PROJECT`

`GOOGLE_APPLICATION_CREDENTIALS`

##### Command:
    
    make gcp
    
#### Azure

##### Prerequisites

**Install [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)**

    az login
    az configure --defaults group=ResourceGroup
    az account set -s SubscriptionID
    
**Create Azure storage account**

    az storage account create -n account -g ResourceGroup
    az storage account keys list -n account -g ResourceGroup

**Set environment variables**

`ARM_STORAGE_ACCOUNT`

`ARM_STORAGE_ACCESS_KEY`

`ARM_SUBSCRIPTION_ID`

`ARM_RES_GROUP_NAME`

`ARM_PROVIDER_VMSS_EXTENSIONS_BETA=true`

**In case you use service principal set next environment variables**:

`ARM_TENANT_ID`

`ARM_CLIENT_ID`

`ARM_CLIENT_SECRET`

`AZURE_TENANT_ID`

`AZURE_CLIENT_ID`

`AZURE_CLIENT_SECRET`

##### Command:
    
    make azure

### All tests

**Set up environment for all projects as described above**

##### Command:

    make all
