@echo off
REM BusinessOS Claude Code Optimization Installer (Windows)
REM Instala Skills, Agents, Hooks e MCP servers automaticamente

setlocal enabledelayedexpansion

echo ╔══════════════════════════════════════════════════════════════╗
echo ║  BusinessOS Claude Code Optimization Installer               ║
echo ╚══════════════════════════════════════════════════════════════╝
echo.

REM Verificar se estamos no diretório correto
if not exist "CLAUDE.md" (
    echo ❌ Erro: Execute este script da raiz do projeto BusinessOS
    exit /b 1
)

echo 📂 Criando estrutura de pastas...
mkdir .claude\skills\go-backend-expert 2>nul
mkdir .claude\skills\svelte-frontend-expert 2>nul
mkdir .claude\skills\database-migration-expert 2>nul
mkdir .claude\skills\testing-expert 2>nul
mkdir .claude\agents 2>nul
mkdir .claude\hooks 2>nul
echo ✅ Estrutura criada
echo.

echo 🎯 Criando Skills e Agents...
echo (Arquivos sendo criados...)

REM Nota: Os arquivos reais devem ser criados manualmente ou através do script bash
REM Este é um wrapper para Windows que informa o usuário

echo.
echo ╔══════════════════════════════════════════════════════════════╗
echo ║  ⚠️  ATENÇÃO - Windows                                       ║
echo ╚══════════════════════════════════════════════════════════════╝
echo.
echo Para Windows, use uma das opções:
echo.
echo OPÇÃO 1 - Git Bash:
echo   1. Abra Git Bash
echo   2. cd C:\Users\Pichau\Desktop\BusinessOS-main-dev
echo   3. bash .claude/install-optimization.sh
echo.
echo OPÇÃO 2 - WSL (Ubuntu):
echo   1. wsl
echo   2. cd /mnt/c/Users/Pichau/Desktop/BusinessOS-main-dev
echo   3. bash .claude/install-optimization.sh
echo.
echo OPÇÃO 3 - Manual:
echo   1. Consulte: docs\CLAUDE_CODE_OPTIMIZATION_GUIDE.md
echo   2. Copie os arquivos manualmente conforme o guia
echo.

pause
