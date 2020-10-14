polkadot tests
=================

#### Prerequisites

* Install **terraform** >= 13.4
* Install **go** >= 1.15.2
* Install **jq** utility 

### Common non-mandatory environment variables

    `PREFIX`            - prefix for TF resources
    `TF_STATE_BUCKET`   - TF state backet name
    `TF_STATE_KEY`      - TF state backet key

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

`AZURE_STORAGE_ACCOUNT`

`AZURE_STORAGE_ACCESS_KEY`

`AZURE_SUBSCRIPTION_ID`

`AZURE_RES_GROUP_NAME`

`AZURE_TENANT_ID`

`AZURE_CLIENT_ID`

`AZURE_CLIENT_SECRET`

##### Command:
    
    make azure

### All tests

**Set up environment for all projects as described above**

##### Command:

    make all
