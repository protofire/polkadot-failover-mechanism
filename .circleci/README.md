# Circle CI configuration

CI pipeline implements several tests against repository code to make sure it can run correctly. It is includes:

- Static Terraform validation
- Functional tests (written on Go)

Upon successful testing the Build phase is triggered. It does:

- Builds docker image, which can be used to deploy the solution
- Makes authentication against DockerHub
- Pushes built image into DockerHub

Please configure following Environment Variables in the Project setting to allow CI tasks to run:

- `AWS_ACCESS_KEY` & `AWS_SECRET_KEY` - To deploy solution to AWS
- `SLACK_WEBHOOK` - To post CI status notifications
- `dockerhub_repo` - Your DockerHub repository to push the image
- `dockerhub_user` - Your DockerHub user
- `dockerhub_token` - Your DockerHub token

Also, since solution is using AWS S3 service to keep Terraform state, you'll need to create a bucket specified in [backends](../aws/backends/s3.tf) config file.