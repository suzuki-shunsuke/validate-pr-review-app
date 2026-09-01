# Logging, Monitoring, Security, etc

Please see the following documents:

- AWS Lambda
  - [Sending Lambda function logs to CloudWatch Logs](https://docs.aws.amazon.com/lambda/latest/dg/monitoring-cloudwatchlogs.html)
  - [Monitoring, debugging, and troubleshooting Lambda functions](https://docs.aws.amazon.com/lambda/latest/dg/lambda-monitoring.html)
- GitHub App
  - [Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)

The log format of Validate PR Review App is the JSON format.
The log has the log level like `INFO`, `WARN`, and `ERROR`, so you can send alerts based on the log level.

<details>
<summary>Example Log</summary>

```json
{
    "time": "2025-09-25T19:49:28.295812986Z",
    "level": "INFO",
    "msg": "Fetched a pull request",
    "version": "",
    "pull_request": {
        "sha": "e21cc0e643655273c71f1d14e3f42ee14c2c6721",
        "approvers": {
            "suzuki-shunsuke": {}
        },
        "commits": [
            {
                "oid": "25031c1e9c20594e0dc4569e193ad3f45de0ade1",
                "committer": {
                    "login": "renovate[bot]",
                    "is_app": true
                },
                "signature": {
                    "isValid": true,
                    "state": "VALID"
                }
            }
        ]
    }
}
```

</details>
