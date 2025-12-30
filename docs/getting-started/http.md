# Getting Stared - HTTP Server

Run HTTP Server in your localhost and receives webhook via smee.io.

Requirements:

- Git
- Docker

1. Checkout the repository

```sh
git clone https://github.com/suzuki-shunsuke/validate-pr-review-app
```

```sh
cd validate-pr-review-app/example
```

2. Pull Docker Image

```
VERSION=v0.1.0-0
```

```sh
docker pull "ghcr.io/suzuki-shunsuke/validate-pr-review-app:$VERSION"
```

3. [Create a GitHub App](#github-app-settings)
4. Prepare config.yaml and secret.yaml
    1. [config.yaml](#non-secret-config)
    2. [secret.yaml](#secrets)

```sh
cp config.yaml.tmpl config.yaml
cp secret.yaml.tmpl secret.yaml
vi config.yaml
vi secret.yaml
```

5. Run the server

```sh
docker run --rm -d -p 8080:8080 \
  -v "$PWD:/workspace" \
  -e "CONFIG_FILE=/workspace/config.yaml" \
  -e "SECRET_FILE=/workspace/secret.yaml" \
  "ghcr.io/suzuki-shunsuke/validate-pr-review-app:$VERSION"
```

6. Receive GitHub Webhook using [smee.io](https://smee.io/)

[See also the GitHub Document `Handling webhook deliveries`](https://docs.github.com/en/webhooks/using-webhooks/handling-webhook-deliveries)

```sh
smee -u <Webhook Proxy URL> -p 8000 --path /webhook
```

7. Create pull requests and reviews
