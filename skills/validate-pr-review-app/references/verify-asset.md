# Verify downloaded assets from GitHub Releases

You can verify downloaded assets using some tools.

1. [GitHub CLI](https://cli.github.com/)
1. [Cosign](https://github.com/sigstore/cosign)
1. [Packslip](https://packslip.dev/) `validate-pr-review-app >= v0.3.4`
1. [slsa-verifier](https://github.com/slsa-framework/slsa-verifier) `validate-pr-review-app <= v0.3.3`

## 1. GitHub CLI

You can install GitHub CLI by aqua.

```sh
aqua g -i cli/cli
```

```sh
version=v0.3.4
asset=validate-pr-review-app_linux_amd64.zip
gh release download -R suzuki-shunsuke/validate-pr-review-app "$version" -p "$asset"
gh attestation verify "$asset" \
  -R suzuki-shunsuke/validate-pr-review-app \
  --signer-workflow suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
```

## 2. Cosign

You can install Cosign by aqua.

```sh
aqua g -i sigstore/cosign
```

From validate-pr-review-app v0.3.3, the bundle file is `validate-pr-review-app_<version>_checksums.txt.sigstore.json`.
Cosign v2.4.2 or later is required.

```sh
version=v0.3.4
checksum_file="validate-pr-review-app_${version#v}_checksums.txt"
asset=validate-pr-review-app_linux_amd64.zip
gh release download "$version" \
  -R suzuki-shunsuke/validate-pr-review-app \
  -p "$asset" \
  -p "$checksum_file" \
  -p "${checksum_file}.sigstore.json"
cosign verify-blob \
  "$checksum_file" \
  --bundle "${checksum_file}.sigstore.json" \
  --certificate-identity-regexp 'https://github\.com/suzuki-shunsuke/go-release-workflow/\.github/workflows/release\.yaml@.*' \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
cat "$checksum_file" | sha256sum -c --ignore-missing -
```

For validate-pr-review-app v0.3.2 or older, `*.pem` and `*.sig` are available instead.

```sh
version=v0.3.2
checksum_file="validate-pr-review-app_${version#v}_checksums.txt"
asset=validate-pr-review-app_linux_amd64.zip
gh release download "$version" \
  -R suzuki-shunsuke/validate-pr-review-app \
  -p "$asset" \
  -p "$checksum_file" \
  -p "${checksum_file}.pem" \
  -p "${checksum_file}.sig"
cosign verify-blob \
  --signature "${checksum_file}.sig" \
  --certificate "${checksum_file}.pem" \
  --certificate-identity-regexp 'https://github\.com/suzuki-shunsuke/go-release-workflow/\.github/workflows/release\.yaml@.*' \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "$checksum_file"
cat "$checksum_file" | sha256sum -c --ignore-missing -
```

## 3. Packslip

`validate-pr-review-app >= v0.3.4`

You can install packslip by aqua.

```sh
aqua g -i jdx/packslip
```

```sh
version=v0.3.4
asset=validate-pr-review-app_linux_amd64.zip
gh release download -R suzuki-shunsuke/validate-pr-review-app "$version" -p "$asset" -p packslip.sigstore.json
packslip verify packslip.sigstore.json \
    --pin ps1_bydkkmgpnm4l24wjkxwmpy5c4m \
    --identity-prefix https://github.com/suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml@ \
    --artifact "$asset"
```

- `--pin` requires the release to be signed from the repository `suzuki-shunsuke/validate-pr-review-app`, identified by its repository ID rather than its name
- `--identity-prefix` requires the release to be signed by the release workflow of [suzuki-shunsuke/go-release-workflow](https://github.com/suzuki-shunsuke/go-release-workflow)

On success, packslip prints a line starting with `ok:`.
Please make sure it shows the project and version you meant, e.g. `github.com/suzuki-shunsuke/validate-pr-review-app 0.3.4`.

## 4. slsa-verifier

> [!WARNING]
> As of validate-pr-review-app v0.3.4, `multiple.intoto.jsonl` hasn't been published anymore.

You can install slsa-verifier by aqua.

```sh
aqua g -i slsa-framework/slsa-verifier
```

```sh
version=v0.3.3
asset=validate-pr-review-app_linux_amd64.zip
gh release download -R suzuki-shunsuke/validate-pr-review-app "$version" -p "$asset" -p multiple.intoto.jsonl
slsa-verifier verify-artifact "$asset" \
  --provenance-path multiple.intoto.jsonl \
  --source-uri github.com/suzuki-shunsuke/validate-pr-review-app \
  --source-tag "$version"
```
