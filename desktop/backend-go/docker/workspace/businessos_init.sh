#!/bin/zsh
# BusinessOS Terminal Init Script
# This file is sourced when the terminal starts

# OSA CLI - 21-Agent Orchestration System
osa() {
  local API_BASE="${BUSINESSOS_API_URL:-http://localhost:8001}"
  local USER_ID="${BUSINESSOS_USER_ID}"

  # Check if user ID is available
  if [ -z "$USER_ID" ]; then
    echo "❌ Error: Not authenticated (BUSINESSOS_USER_ID not set)"
    return 1
  fi

  case "$1" in
    health)
      echo "🔍 Checking OSA-5 health..."
      curl -s -H "X-User-ID: $USER_ID" "$API_BASE/api/internal/osa/health" | jq -r 'if .status == "healthy" then "✅ OSA-5 is healthy\n   Status: \(.status)\n   Version: \(.version)" else "❌ OSA-5 is \(.status)" end'
      ;;

    generate|gen)
      if [ -z "$2" ]; then
        echo "Usage: osa generate <description>"
        echo "Example: osa generate \"task management system with kanban board\""
        return 1
      fi

      local description="${*:2}"
      echo "🎯 Generating BusinessOS module: $description"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

      local response=$(curl -s "$API_BASE/api/internal/osa/generate" \
        -H "Content-Type: application/json" \
        -H "X-User-ID: $USER_ID" \
        -d "{\"name\":\"Generated Module\",\"description\":\"$description\",\"type\":\"fullstack\"}" 2>&1)

      local app_id=$(echo "$response" | jq -r '.app_id // .appId // empty' 2>/dev/null)

      if [ -n "$app_id" ]; then
        echo "✅ Generation started!"
        echo "   App ID: $app_id"
        echo ""
        echo "⏳ OSA-5 is running 21-agent workflow..."
        echo "   Use 'osa status $app_id' to check progress"
      else
        echo "❌ Generation failed:"
        local error=$(echo "$response" | jq -r '.error // .details // empty' 2>/dev/null)
        if [ -z "$error" ]; then
          echo "   OSA-5 API is not fully configured yet"
          echo "   The service is running but endpoints need to be set up"
        else
          echo "   $error"
        fi
      fi
      ;;

    status)
      if [ -z "$2" ]; then
        echo "Usage: osa status <app-id>"
        return 1
      fi

      echo "📊 Checking status for: $2"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

      curl -s -H "X-User-ID: $USER_ID" "$API_BASE/api/internal/osa/status/$2" | jq -r '
        if .status then
          "App ID:       \(.app_id // .appId)\n" +
          "Status:       \(.status)\n" +
          "Progress:     \((.progress // 0) * 100)%\n" +
          (if .current_step then "Current Step: \(.current_step)\n" else "" end) +
          (if .error then "\n❌ Error: \(.error)\n" else "" end)
        else
          "❌ Error: " + (.error // .message // "Unknown error")
        end'
      ;;

    list)
      echo "📦 Recent Apps"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      curl -s -H "X-User-ID: $USER_ID" "$API_BASE/api/internal/osa/workspaces" | jq -r '.workspaces[] | "  • \(.name) (\(.app_count) apps)"'
      ;;

    help|--help|-h)
      cat << 'EOF'
OSA CLI - Control the 21-Agent Orchestration System
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Commands:
  osa generate <description>  Generate a new BusinessOS module
  osa gen <description>        Alias for 'generate'
  osa status <app-id>          Check module generation status
  osa list                     List workspaces and apps
  osa health                   Check OSA-5 health
  osa help                     Show this help

Examples:
  osa gen "expense tracking with receipts"
  osa gen "inventory management system"
  osa status app-abc-123
  osa health

Note: Generates modules WITHIN BusinessOS (not standalone apps)
      Generated code follows BusinessOS patterns (Svelte + Go + PostgreSQL)

Documentation: https://docs.businessos.ai/osa-integration
EOF
      ;;

    *)
      echo "Unknown command: $1"
      echo "Try 'osa help' for available commands"
      return 1
      ;;
  esac
}

# Welcome message
echo ""
echo "🎯 BusinessOS Terminal Ready"
echo "   Type 'osa help' to generate modules with AI"
echo ""
