set -euo pipefail
          mkdir -p "${OUTDIR}"
          build() {
            local goos="$1" goarch="$2"
            local ext=""
            [[ "$goos" == "windows" ]] && ext=".exe"
            local out="${OUTDIR}/${BINARY_NAME}_${VERSION}_${goos}_${goarch}${ext}"
            echo "-> ${goos}/${goarch} -> ${out}"
            CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" GOEXPERIMENT=boringcrypto \
              go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
              -o "${out}" "${PKG}"
          }
          build linux amd64
          build linux arm64
          build darwin amd64
          build darwin arm64
          build windows amd64
          cd "${OUTDIR}"
          for f in ${BINARY_NAME}_*; do
            base="${f%.*}"
            if [[ "$f" == *.exe ]]; then
              zip "${base}.zip" "$f"
            else
              tar -czf "${base}.tar.gz" "$f"
            fi
          done