#!/usr/bin/env bash
set -euo pipefail

IMAGE="${IMAGE:-tg-bridge:latest}"

BUILDER="${BUILDER:-paketobuildpacks/builder-jammy-buildpackless-static}"
BUILDPACK="${BUILDPACK:-paketo-buildpacks/go}"

RUN_IMAGE="${RUN_IMAGE:-}"

PUBLISH="${PUBLISH:-false}"

echo "📦 Image:   ${IMAGE}"
echo "🚧 Builder: ${BUILDER}"
echo "🥃 Buildpack: ${BUILDPACK}"
echo "✈️ Publish: ${PUBLISH}"

ARGS=(build "${IMAGE}" --builder "${BUILDER}" --buildpack "${BUILDPACK}" --pull-policy always)
if [[ "${PUBLISH}" == "true" ]]; then
  ARGS+=(--publish)
fi

pack "${ARGS[@]}"
echo "Done: ${IMAGE}"