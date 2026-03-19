# Getting Stared - HTTP Server

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

2. [Create a GitHub App](../github-app.md)
3. Prepare config.yaml and secret.yaml
    1. [config.yaml](../config.md#non-secret-config)
    2. [secret.yaml](../config.md#secrets)

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
export CONFIG_FILE=example/config.yaml
export SECRET_FILE=example/secret.yaml
go run ./cmd/app
```

5. Receive GitHub Webhook using [smee.io](https://smee.io/)

[See also the GitHub Document `Handling webhook deliveries`](https://docs.github.com/en/webhooks/using-webhooks/handling-webhook-deliveries)

```sh
smee -u <Webhook Proxy URL> -p 8080 --path /webhook
```

6. Create pull requests and reviews
