# Circle CI configuration

CI pipeline implements several tests against repository code to make sure it can run correctly. It is includes:

- Static Terraform validation
- Functional tests (written on Go)

Upon successful testing the Build phase is triggered. It does:

- Builds docker image, which can be used to deploy the solution
- Makes authentication against DockerHub
- Pushes built image into DockerHub

Please configure following Environment Variables in the Project setting to allow CI tasks to run:

- `TF_STATE_BUCKET` & `TF_STATE_KEY` - For all cloud providers
- `AWS_ACCESS_KEY` & `AWS_SECRET_KEY` - To deploy solution to AWS
- `GCP_PROJECT` & `GOOGLE_APPLICATION_CREDENTIALS_CONTENT` & `GOOGLE_APPLICATION_CREDENTIALS` - To deploy solution to GCP 
-   `ARM_SUBSCRIPTION_ID` & 
    `ARM_TENANT_ID` &
    `ARM_CLIENT_ID` &
    `ARM_CLIENT_SECRET` &
    `AZURE_TENANT_ID` &
    `AZURE_CLIENT_ID` &
    `AZURE_CLIENT_SECRET` &
    `ARM_RES_GROUP_NAME` &
    `ARM_STORAGE_ACCOUNT` &
    `ARM_STORAGE_ACCESS_KEY` &
    `ARM_PROVIDER_VMSS_EXTENSIONS_BETA` - To deploy solution to Azure

- `SLACK_WEBHOOK` - To post CI status notifications
- `dockerhub_repo` - Your DockerHub repository to push the image
- `dockerhub_user` - Your DockerHub user
- `dockerhub_token` - Your DockerHub token
