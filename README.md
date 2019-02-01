# mario-powerup :mushroom:

mario-powerup is a Lambda function that starts the mario Fargate task. There's no easy way to trigger a Fargate task from an S3 event while passing the uploaded file name to the task. This function serves as the intermediary.

The Fargate task is initially configured in the Terraform config for mario, however some of the task configurations need to be passed again here in the Lambda for it to work. Those config items that are needed are set as environment variables on the Lambda by Terraform when it is created.

# Deploying

Staging builds are handled automatically by Travis on a PR merge. Deploying the staging build to production will require the manual step of running `make promote`. This copies the `mario.zip` deployment package from the staging S3 bucket to the production S3 bucket and then updates the Lambda function code.
