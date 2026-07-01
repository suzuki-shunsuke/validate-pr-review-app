#!/usr/bin/env bash

set -euo pipefail

TAG=$1

IMAGE=ghcr.io/suzuki-shunsuke/validate-pr-review-app:$TAG
docker pull $IMAGE
repo_digest=$(docker inspect --format='{{index .RepoDigests 0}}' $IMAGE)
digest=${repo_digest##*@}

sign_digest=$(cosign verify \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com \
    --certificate-identity "https://github.com/suzuki-shunsuke/validate-pr-review-app/.github/workflows/release.yaml@refs/tags/$TAG" \
    --certificate-github-workflow-ref "refs/tags/$TAG" \
    --certificate-github-workflow-repository "suzuki-shunsuke/validate-pr-review-app" \
    $IMAGE |
    jq -r '.[].critical.image."docker-manifest-digest"')

if [ "$digest" = "$sign_digest" ]; then
    echo "[INFO] Signature verification succeeded: $IMAGE" >&2
else
    echo "[ERROR] Signature verification failed: $IMAGE" >&2
    echo "Expected digest: $sign_digest, Actual digest: $digest" >&2
    exit 1
fi
