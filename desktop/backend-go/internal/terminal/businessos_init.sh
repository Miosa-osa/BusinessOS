#!/bin/bash
# BusinessOS Terminal Init Script
# This file is sourced when the terminal starts (works in both bash and zsh)

# OSA CLI - 21-Agent Orchestration System
osa() {
  # Detect if running in container
  local IN_CONTAINER=false
  if [ -f /.dockerenv ] || grep -q docker /proc/1/cgroup 2>/dev/null; then
    IN_CONTAINER=true
  fi

  # API endpoints - use host.docker.internal in containers
  local BACKEND_API="http://localhost:8001"
  local OSA_API="http://localhost:3003"
  if [ "$IN_CONTAINER" = true ]; then
    BACKEND_API="http://host.docker.internal:8001"
    OSA_API="http://host.docker.internal:3003"
  fi

  local USER_ID="${BUSINESSOS_USER_ID}"

  # Check if user ID is available
  if [ -z "$USER_ID" ]; then
    echo "[ERROR] Not authenticated (BUSINESSOS_USER_ID not set)"
    return 1
  fi

  case "$1" in
    ""|chat)
      # Interactive chat mode
      echo ""
      echo "============================================================"
      echo "  OSA INTERACTIVE MODE"
      echo "  Type your requests, 'exit' to quit"
      echo "============================================================"
      echo ""

      while true; do
        # Read user input
        printf "> "
        read -r user_input

        # Check for exit
        case "$user_input" in
          exit|quit|q)
            echo ""
            echo "Goodbye!"
            echo ""
            return 0
            ;;
          "")
            continue
            ;;
          help|h|\?)
            echo ""
            echo "Commands:"
            echo "  Type any request to generate code"
            echo "  'exit' or 'quit' - Exit chat"
            echo "  'clear' - Clear screen"
            echo ""
            continue
            ;;
          clear)
            clear
            echo "OSA Interactive Mode (type 'exit' to quit)"
            echo ""
            continue
            ;;
        esac

        echo ""
        echo "Processing..."
        echo ""

        # Call the orchestrator
        local response=$(curl -s "$OSA_API/api/orchestrate" \
          -H "Content-Type: application/json" \
          -d "{\"prompt\":\"$user_input\",\"context\":{\"user_id\":\"$USER_ID\",\"platform\":\"businessos\"}}" 2>&1)

        # Check if successful and display nicely
        if echo "$response" | grep -q '"success":true'; then
          echo "------------------------------------------------------------"
          echo "  RESPONSE"
          echo "------------------------------------------------------------"
          echo ""

          # Extract and display the output text from each agent
          # Get the last agent's output (most relevant)
          local raw_output=$(echo "$response" | grep -oE '"output":"[^"]*"' | tail -1 | cut -d'"' -f4)

          # Convert escape sequences properly
          local output=$(printf '%b' "$raw_output" | head -50)

          if [ -n "$output" ]; then
            printf '%s\n' "$output" | head -40
            local lines=$(echo "$output" | wc -l)
            if [ "$lines" -gt 40 ]; then
              echo ""
              echo "... (truncated, $lines total lines)"
            fi
          else
            # Fallback - show summary
            local agents=$(echo "$response" | grep -o '"next_agent":"[^"]*"' | wc -l | tr -d ' ')
            local files=$(echo "$response" | grep -oE '"path":"[^"]*"' | cut -d'"' -f4 | head -5)
            echo "  Agents completed: $agents"
            if [ -n "$files" ]; then
              echo "  Files generated:"
              echo "$files" | while read -r f; do
                [ -n "$f" ] && echo "    • $f"
              done
            fi
          fi
          echo ""
        else
          echo "------------------------------------------------------------"
          echo "  ERROR"
          echo "------------------------------------------------------------"
          echo ""
          local error=$(echo "$response" | grep -oE '"error":"[^"]*"' | head -1 | cut -d'"' -f4)
          if [ -n "$error" ]; then
            echo "  $error"
          else
            echo "  Request failed. Try again."
          fi
          echo ""
        fi
      done
      ;;

    health)
      echo "Checking OSA health..."
      local result=$(curl -s "$OSA_API/health" 2>&1)
      if [ "$result" = "OK" ]; then
        echo "[OK] OSA is healthy"
        echo "     Orchestrator: $OSA_API"
      else
        echo "[ERROR] OSA is not responding"
        echo "        Response: $result"
      fi
      ;;

    agents)
      echo "Available Agents"
      echo "=============================================="
      local result=$(curl -s "$OSA_API/api/agents" 2>&1)
      if command -v jq >/dev/null 2>&1; then
        echo "$result" | jq -r '.[] | "  • \(.type): \(.capabilities | join(", "))"' 2>/dev/null || echo "Response: $result"
      else
        echo "$result"
      fi
      ;;

    generate|gen)
      if [ -z "$2" ]; then
        echo "Usage: osa generate <description>"
        echo "Example: osa generate \"task management system with kanban board\""
        return 1
      fi

      local description="${*:2}"
      echo "Generating BusinessOS module: $description"
      echo "=============================================="
      echo ""
      echo "Running 21-agent orchestration..."
      echo ""

      # Call the orchestrator directly
      local response=$(curl -s "$OSA_API/api/orchestrate" \
        -H "Content-Type: application/json" \
        -d "{\"prompt\":\"$description\",\"context\":{\"user_id\":\"$USER_ID\",\"platform\":\"businessos\"}}" 2>&1)

      # Check if successful
      if echo "$response" | grep -q '"success":true'; then
        echo ""
        echo "============================================================"
        echo "  GENERATION COMPLETE"
        echo "============================================================"
        echo ""

        # Extract workflow ID
        local wf_id=$(echo "$response" | grep -oE '"workflow_id":"[^"]+"' | head -1 | cut -d'"' -f4)
        echo "  Workflow: ${wf_id:-N/A}"
        echo ""

        # Count successful agents
        local success_count=$(echo "$response" | grep -o '"success":true' | wc -l | tr -d ' ')
        echo "  AGENTS EXECUTED: $success_count"
        echo "  ----------------------------"

        # Extract agent info using grep/sed
        local agents=$(echo "$response" | grep -oE '"next_agent":"[^"]*"' | cut -d'"' -f4)
        local times=$(echo "$response" | grep -oE '"execution_ms":[0-9]+' | cut -d':' -f2)
        local confidences=$(echo "$response" | grep -oE '"confidence":[0-9.]+' | cut -d':' -f2)

        # Display each agent
        local i=1
        echo "$agents" | while read -r agent; do
          local time=$(echo "$times" | sed -n "${i}p")
          local conf=$(echo "$confidences" | sed -n "${i}p")
          printf "  %d. %-12s  %4sms  %.0f%% confidence\n" "$i" "${agent:-analysis}" "${time:-0}" "$(echo "${conf:-0} * 100" | bc 2>/dev/null || echo 0)"
          i=$((i + 1))
        done

        echo ""

        # Extract and show generated files
        local files=$(echo "$response" | grep -oE '"path":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$files" ]; then
          echo "  GENERATED FILES:"
          echo "  -----------------"
          echo "$files" | while read -r file; do
            [ -n "$file" ] && echo "    - $file"
          done
        fi

        echo ""
        echo "  Code generated! Check the OSA workspace for files."
        echo ""
      else
        echo ""
        echo "============================================================"
        echo "  GENERATION FAILED"
        echo "============================================================"
        echo ""
        local error=$(echo "$response" | grep -oE '"error":"[^"]*"' | head -1 | cut -d'"' -f4)
        if [ -n "$error" ]; then
          echo "  Error: $error"
        else
          echo "  Unknown error occurred"
        fi
        echo ""
      fi
      ;;

    status)
      if [ -z "$2" ]; then
        echo "Usage: osa status <app-id>"
        return 1
      fi

      echo "Checking status for: $2"
      echo "=============================================="

      local result=$(curl -s -H "X-User-ID: $USER_ID" "$BACKEND_API/api/internal/osa/status/$2" 2>&1)
      if command -v jq >/dev/null 2>&1; then
        echo "$result" | jq -r '
          if .status then
            "App ID:       \(.app_id // .appId)\n" +
            "Status:       \(.status)\n" +
            "Progress:     \((.progress // 0) * 100)%\n" +
            (if .current_step then "Current Step: \(.current_step)\n" else "" end) +
            (if .error then "\n[ERROR] \(.error)\n" else "" end)
          else
            "[ERROR] " + (.error // .message // "Unknown error")
          end'
      else
        echo "Response: $result"
      fi
      ;;

    list)
      echo "This command is not yet available."
      echo "Use 'osa agents' to see available agents."
      echo "Use 'osa gen <description>' to generate a module."
      ;;

    help|--help|-h)
      cat << 'EOF'
OSA CLI - 21-Agent AI Orchestration System
==============================================

Commands:
  osa                         Start interactive chat mode
  osa chat                    Alias for interactive mode
  osa gen <description>       Quick one-shot generation
  osa agents                  List available AI agents
  osa health                  Check OSA health
  osa help                    Show this help

Interactive Mode:
  Just type 'osa' to start chatting with OSA.
  Type your requests and get AI-generated responses.
  Type 'exit' to leave interactive mode.

Examples:
  osa                         # Start chatting
  osa gen "task manager"      # Quick generation
  osa health                  # Check status

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
echo "BusinessOS Terminal Ready"
echo "Type 'osa help' to generate modules with AI"
echo ""
