---
name: validate-pr-review-app-verify-assets
description: |
  Verify the authenticity of validate-pr-review-app release assets and container images
  downloaded from GitHub Releases. Covers GitHub CLI attestation verification, slsa-verifier with
  SLSA provenance, and Cosign for both release-asset checksums and container images.
  Use this skill when the user wants to verify a downloaded binary, release asset, checksum, or
  container image, or set up supply-chain verification of the app.
---

Read [reference.md](reference.md) to verify release assets and container images — using GitHub
CLI attestation, slsa-verifier (SLSA provenance), or Cosign, including verifying container
images with the bundled `verify-image.sh`.
