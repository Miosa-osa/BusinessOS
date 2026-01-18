#!/bin/bash
# Generate Go code from protobuf definitions

set -e

# Add Go bin to PATH for protoc plugins
export PATH="$PATH:$HOME/go/bin"

# Change to project root
cd "$(dirname "$0")/.."

echo "🔧 Generating Go code from proto files..."

# Generate Go code for voice service
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/voice/v1/voice.proto

echo "✅ Proto generation complete!"
echo ""
echo "Generated files:"
echo "  - proto/voice/v1/voice.pb.go (message types)"
echo "  - proto/voice/v1/voice_grpc.pb.go (gRPC service)"
