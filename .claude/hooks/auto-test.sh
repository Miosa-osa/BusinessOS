#!/bin/bash
FILE=$(echo "$1" | jq -r '.tool_input.file_path')

# Se modificou arquivo Go, roda testes do pacote
if [[ "$FILE" == *.go ]] && [[ "$FILE" != *_test.go ]]; then
  DIR=$(dirname "$FILE")
  echo "🧪 Running tests for $DIR..."
  cd desktop/backend-go && go test "./$DIR" -short 2>&1 | head -20
fi

# Se modificou arquivo Svelte/TS, roda testes relacionados
if [[ "$FILE" == *.svelte ]] || [[ "$FILE" == *.ts ]]; then
  echo "🧪 Running related tests..."
  cd frontend && npm test -- "$FILE" 2>&1 | head -20
fi
