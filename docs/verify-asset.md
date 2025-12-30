# Verify downloaded assets from GitHub Releases

You can verify downloaded assets using some tools.

1. [GitHub CLI](https://cli.github.com/)
1. [slsa-verifier](https://github.com/slsa-framework/slsa-verifier)
1. [Cosign](https://github.com/sigstore/cosign)

## 1. GitHub CLI

You can install GitHub CLI by aqua.

```sh
aqua g -i cli/cli
```

```sh
version=v1.0.0
asset=validate-pr-review-app_linux_amd64.zip
gh release download -R suzuki-shunsuke/validate-pr-review-app "$version" -p "$asset"
gh attestation verify "$asset" \
  -R suzuki-shunsuke/validate-pr-review-app \
  --signer-workflow suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
```

## 2. slsa-verifier

You can install slsa-verifier by aqua.

```sh
aqua g -i slsa-framework/slsa-verifier
```

```sh
version=v1.0.0
asset=validate-pr-review-app_linux_amd64.zip
gh release download -R suzuki-shunsuke/validate-pr-review-app "$version" -p "$asset" -p multiple.intoto.jsonl
slsa-verifier verify-artifact "$asset" \
  --provenance-path multiple.intoto.jsonl \
  --source-uri github.com/suzuki-shunsuke/validate-pr-review-app \
  --source-tag "$version"
```

## 3. Cosign

You can install Cosign by aqua.

```sh
aqua g -i sigstore/cosign
```

```sh
version=v1.0.0
checksum_file="validate-pr-review-app_${version#v}_checksums.txt"
asset=validate-pr-review-app_linux_amd64.zip
gh release download "$version" \
  -R suzuki-shunsuke/validate-pr-review-app \
  -p "$asset" \
  -p "$checksum_file" \
  -p "${checksum_file}.bundle"
cosign verify-blob \
  --bundle "${checksum_file}.bundle" \
  --certificate-identity-regexp 'https://github\.com/suzuki-shunsuke/go-release-workflow/\.github/workflows/release\.yaml@.*' \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "$checksum_file"
cat "$checksum_file" | sha256sum -c --ignore-missing
```
