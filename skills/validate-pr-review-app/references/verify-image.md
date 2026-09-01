# Verify Container Images

You can verify container images using Cosign.

You can install Cosign by aqua.

```sh
aqua g -i sigstore/cosign
```

[verify-image.sh](https://github.com/suzuki-shunsuke/validate-pr-review-app/blob/main/scripts/verify-image.sh)

```sh
bash scripts/verify-image.sh v0.1.0-0
```
