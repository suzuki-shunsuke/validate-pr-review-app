# Getting Started - HTTP Server

Run HTTP Server in your localhost and receives webhook via smee.io.

Requirements:

- Git
- Docker
  - Instead of Docker, You can also run the app from source code using Go.

1. Checkout the repository

```sh
git clone https://github.com/suzuki-shunsuke/validate-pr-review-app
```

```sh
cd validate-pr-review-app/example
```

2. [Create a GitHub App](github-app.md)
3. Prepare config.yaml and secret.yaml
    1. [config.yaml](config.md)
    2. [secret.yaml](secret.md)

```sh
cp config.yaml.tmpl config.yaml
cp secret.yaml.tmpl secret.yaml
vi config.yaml
vi secret.yaml
```

4. Run the server

You can run the server using Docker or Go.

Docker:

```sh
VERSION=v0.1.0-0
docker run --rm -d -p 8080:8080 \
  -v "$PWD:/workspace" \
  -e "CONFIG_FILE=/workspace/config.yaml" \
  -e "SECRET_FILE=/workspace/secret.yaml" \
  "ghcr.io/suzuki-shunsuke/validate-pr-review-app:$VERSION"
```

Go:

```sh
bash start-go.sh
```

Or

```sh
export CONFIG_FILE=example/config.yaml
export SECRET_FILE=example/secret.yaml
go run ./cmd/app
```

5. Receive GitHub Webhook using [smee.io](https://smee.io/) ([smee-client](https://www.npmjs.com/package/smee-client))

[See also the GitHub Document `Handling webhook deliveries`](https://docs.github.com/en/webhooks/using-webhooks/handling-webhook-deliveries)

```sh
smee -u <Webhook Proxy URL> -p 8080 --path /webhook
```

6. Change the GitHub App Webhook URL to `<Webhook Proxy URL>`.
7. Create pull requests and reviews

## Endpoints

- `GET /webhook`: GitHub Webhook Handler
- `GET /ready`: Response 200 `{"status": "ok"}`

`/webhook` returns 200 immediately and processes the webhook event in the background, because GitHub treats a delivery as failed if the server doesn't respond within 10 seconds.
So the response status doesn't reflect the processing result. Check the server logs instead.
On shutdown, the server waits for webhook events being processed in the background, and returns 503 to new webhook requests so that GitHub treats them as failed.
GitHub doesn't redeliver failed deliveries automatically, but you can redeliver them manually from the GitHub App's Recent Deliveries.
The server doesn't limit how long it waits, so set the platform's grace period (e.g. `terminationGracePeriodSeconds` on Kubernetes) longer than the worst-case processing time.
Otherwise the process is killed during processing and the Check Run isn't created.
Sending a second SIGINT or SIGTERM terminates the server immediately.
`/webhook` rejects request bodies larger than 25 MB, which is GitHub's cap on webhook payloads.
If you run the server on a platform that throttles CPU after a response is returned (e.g. Cloud Run with request-based billing), allocate CPU always.

## Files

- [example/config.yaml.tmpl](https://github.com/suzuki-shunsuke/validate-pr-review-app/blob/main/example/config.yaml.tmpl)
- [example/secret.yaml.tmpl](https://github.com/suzuki-shunsuke/validate-pr-review-app/blob/main/example/secret.yaml.tmpl)
- [example/start-go.sh](https://github.com/suzuki-shunsuke/validate-pr-review-app/blob/main/example/start-go.sh)

## See Also

- [Environment Variables](env.md)
- [Logging, Monitoring, and Security](production.md)
