#!/bin/bash
# Script para atualizar referências do repositório antigo para o novo
# De: robertohluna/BusinessOS
# Para: Miosa-osa/BusinessOS

echo "🔄 Atualizando referências do repositório..."

# 1. Atualizar git remote
echo "📍 Atualizando git remote..."
git remote set-url origin https://github.com/Miosa-osa/BusinessOS.git
git remote -v

# 2. Atualizar README.md
echo "📝 Atualizando README.md..."
sed -i 's|github.com/robertohluna/BusinessOS|github.com/Miosa-osa/BusinessOS|g' README.md

# 3. Verificar se há outras referências
echo "🔍 Procurando outras referências..."
grep -r "robertohluna" --include="*.md" --include="*.json" --include="*.yml" --include="*.yaml" . 2>/dev/null || echo "✅ Nenhuma referência encontrada"

echo ""
echo "✅ Atualização concluída!"
echo ""
echo "📋 Arquivos modificados:"
git status --short

echo ""
echo "💡 Próximos passos:"
echo "1. Revisar mudanças: git diff"
echo "2. Commit: git add -A && git commit -m 'chore: update repository URLs to Miosa-osa'"
echo "3. Push: git push"
