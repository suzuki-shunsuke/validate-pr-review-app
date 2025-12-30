# AWS Lambda

Deploying the app to AWS Lambda using Terraform.

Requirements:

- AWS Account
- Terraform
- Git
- bash
- GitHub CLI
- GitHub Repository where the app is installed.
  - A pull request you haven't pushed any commit is necessary

1. Checkout the repository

```sh
git clone https://github.com/suzuki-shunsuke/validate-pr-review-app
```

2. Move to [terraform/aws](../../terraform/aws).

```sh
cd validate-pr-review-app/terraform/aws
```

3. Run `bash init.sh`

```sh
bash init.sh
```

4. [Registering a GitHub App](../github-app.md)

Deactivate Webhook for now. We'll enable this after deploying the AWS Lambda Function.

5. Add the private key to [secret.yaml](../../terraform/aws/secret.yaml.tmpl) and remove the downloaded private key file.

secret.yaml

```yaml
github_app_private_key: |
  -----BEGIN RSA PRIVATE KEY-----
  ...
```

7. [Install the app to your repository](https://docs.github.com/en/apps/using-github-apps/installing-your-own-github-app)

Please install the app to your repository.
[If you don't have any repository for this, please create a repository.](https://github.com/new)

8. [Create a secret token](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries) and add it to secret.yaml

secret.yaml

```yaml
webhook_secret: 0123456789abcdefghijklmn
github_app_private_key: |
  -----BEGIN RSA PRIVATE KEY-----
  ...
```

9. Deploy the app by Terraform

(Optional) If you want to change input variables, please check [variables.tf](../../terraform/aws/variables.tf) and create a file `terraform.tfvars`.

e.g.

```
region = "ap-northeast-1" # Default: us-east-1
```

Then running Terraform commands.

```sh
terraform init
terraform validate
terraform plan
terraform apply
```

10. [Configure Webhook](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/using-webhooks-with-github-apps)

You can check the webhook URL by `terraform state show`.

```console
$ terraform state show 'aws_lambda_function_url.main[0]'
# aws_lambda_function_url.main[0]:
resource "aws_lambda_function_url" "main" {
    # ...
    function_url       = "https://abcdefghijklmnopqrstuvwxyz012345.lambda-url.us-east-1.on.aws/"
    # ...
}
```

- Set the Lambda Function URL to the webhook URL
- Set the secret token to the webhook secret

Please don't forget to click `Save changes`.

11. Subscribe Events for GitHub App

Checks the following events.

- Pull request review

And click `Save changes`.

If the button `Save changes` is disabled and you can't clike `Save changes`, please try to change any permission and revert the change.

The setup was done.
You can create reviews and try the app.
Please prepare a pull request that you haven't pushed any commit.

12. Approve a pull request

Then the check passes.

If the app doesn't work, please check the AWS CloudWatch Log to check if the request reached to the AWS Lambda.

13. Dismiss the review.

Then the check fails.

14. Clean up

You can destroy resources by `terraform destroy`.

```sh
terraform destroy
```

- Uninstall the GitHub App from the repository
- Delete the GitHub App

## Using Amazon API Gateway instead of Lambda Function URL

Amazon API Gateway is also available instead of Lambda Function URL.
ref. [Select a method to invoke your Lambda function using an HTTP request](https://docs.aws.amazon.com/lambda/latest/dg/furls-http-invoke-decision.html).

Remove `use_lambda_function_url` from [config.yaml](../../terraform/aws/config.yaml.tmpl).

```yaml
aws:
  secret_id: validate-pr-review-app
  # use_lambda_function_url: true
```
