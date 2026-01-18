#!/bin/bash
# Generate Python gRPC stubs from proto files

set -e

echo "🔧 Generating Python gRPC stubs..."

# Get the proto file from Go backend
PROTO_FILE="../desktop/backend-go/proto/voice/v1/voice.proto"

if [ ! -f "$PROTO_FILE" ]; then
    echo "❌ Proto file not found: $PROTO_FILE"
    exit 1
fi

# Generate Python code
python3 -m grpc_tools.protoc \
    -I../desktop/backend-go/proto \
    --python_out=. \
    --grpc_python_out=. \
    voice/v1/voice.proto

echo "✅ Python gRPC stubs generated!"
echo ""
echo "Generated files:"
echo "  - voice/v1/voice_pb2.py (message types)"
echo "  - voice/v1/voice_pb2_grpc.py (gRPC service)"
