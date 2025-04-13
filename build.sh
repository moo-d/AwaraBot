#!/bin/bash

OUT_DIR="./bin"
mkdir -p $OUT_DIR

PLATFORMS=(
  "windows/amd64:.exe"
  "linux/amd64:-linux-x64"
  "darwin/amd64:-macos-x64"
  "darwin/arm64:-macos-arm64"
)

for platform in "${PLATFORMS[@]}"; do
  PLATFORM=${platform%:*}
  SUFFIX=${platform#*:}
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}

  echo "Building for $GOOS/$GOARCH..."
  OUTPUT="$OUT_DIR/whatsapp-bot$SUFFIX"
  GOOS=$GOOS GOARCH=$GOARCH go build -o $OUTPUT ./cmd/bot
  
  if [ "$GOOS" != "windows" ]; then
    chmod +x $OUTPUT
  fi
  
  echo "Binary created at $OUTPUT"
done

echo "Build completed. Binaries are in $OUT_DIR"
