# ============================================================================
# Script de Aplicação de Migrations no Supabase
# Autor: Claude Code
# Data: 2026-01-02
# ============================================================================

Write-Host "╔════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║     Assistente de Migrations - Supabase                ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

$projectRoot = Split-Path -Parent $PSScriptRoot
$sqlFile = Join-Path $projectRoot "supabase-migrations-combined.sql"
$supabaseUrl = "https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/sql/new"

# Verificar se o arquivo SQL existe
if (-not (Test-Path $sqlFile)) {
    Write-Host "❌ Arquivo SQL não encontrado: $sqlFile" -ForegroundColor Red
    Write-Host ""
    Write-Host "Execute primeiro: node apply-migrations.js" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ Arquivo SQL encontrado!" -ForegroundColor Green
Write-Host ""

# Copiar conteúdo para clipboard
Write-Host "📋 Copiando conteúdo do SQL para a área de transferência..." -ForegroundColor Yellow
try {
    Get-Content $sqlFile -Raw | Set-Clipboard
    Write-Host "✅ SQL copiado para a área de transferência!" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Não foi possível copiar automaticamente." -ForegroundColor Yellow
    Write-Host "   Você precisará copiar manualmente." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "              INSTRUÇÕES PASSO A PASSO" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

Write-Host "1️⃣  Vou abrir o Supabase SQL Editor no seu navegador" -ForegroundColor White
Write-Host "2️⃣  O SQL já está COPIADO na área de transferência" -ForegroundColor White
Write-Host "3️⃣  Cole (Ctrl+V) no editor SQL do Supabase" -ForegroundColor White
Write-Host "4️⃣  Clique no botão 'RUN' ▶️" -ForegroundColor White
Write-Host "5️⃣  Aguarde a execução (pode levar 1-2 minutos)" -ForegroundColor White
Write-Host ""

Write-Host "⚠️  IMPORTANTE: Habilite a extensão pgvector primeiro!" -ForegroundColor Yellow
Write-Host "   Se aparecer erro de 'vector', execute isto ANTES:" -ForegroundColor Yellow
Write-Host "   CREATE EXTENSION IF NOT EXISTS vector;" -ForegroundColor White
Write-Host ""

Write-Host "Pressione ENTER para abrir o Supabase SQL Editor..." -ForegroundColor Green
$null = Read-Host

# Abrir Supabase SQL Editor
Write-Host ""
Write-Host "🌐 Abrindo Supabase SQL Editor..." -ForegroundColor Yellow
Start-Process $supabaseUrl

Write-Host ""
Write-Host "✅ Pronto! Agora:" -ForegroundColor Green
Write-Host "   1. Cole o SQL (Ctrl+V) no editor" -ForegroundColor White
Write-Host "   2. Clique em RUN ▶️" -ForegroundColor White
Write-Host "   3. Aguarde o sucesso" -ForegroundColor White
Write-Host ""

Write-Host "📝 Se algo der errado, abra o arquivo manualmente:" -ForegroundColor Yellow
Write-Host "   $sqlFile" -ForegroundColor Gray
Write-Host ""

Write-Host "Pressione ENTER após executar as migrations no Supabase..." -ForegroundColor Green
$null = Read-Host

# Verificar resultado
Write-Host ""
Write-Host "🔍 Vou verificar se as migrations foram aplicadas..." -ForegroundColor Yellow
Write-Host ""

Set-Location $projectRoot
& node verify-supabase-schema.js

Write-Host ""
Write-Host "✅ Script finalizado!" -ForegroundColor Green
Write-Host ""
