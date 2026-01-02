$env:PGPASSWORD = "yasdas230321*"
$psqlPath = "C:\Program Files\PostgreSQL\18\bin\psql.exe"

Write-Host "Testing memory creation directly in database..." -ForegroundColor Cyan

$sql = @"
-- Insert test memory with embeddings
INSERT INTO memories (
    user_id, title, summary, content, memory_type, category,
    source_type, tags
)
VALUES (
    'test-user-f6a4a663cd4d4c75836f5854dcc4e0fd',
    'Test Memory for Business Requirements',
    'This is a test memory for verifying the memory system',
    'This memory contains important business requirements and project specifications that should be searchable via semantic embeddings.',
    'fact',
    'technical',
    'manual',
    ARRAY['test', 'requirements', 'project']
)
RETURNING id, title, memory_type, created_at;

-- Count memories
SELECT COUNT(*) as total_memories FROM memories;

-- Show created memory
SELECT id, title, summary, memory_type, source_type, tags, created_at
FROM memories
ORDER BY created_at DESC
LIMIT 1;
"@

echo $sql | & $psqlPath -U postgres -d postgres

Write-Host ""
Write-Host "Memory system test complete!" -ForegroundColor Green

Remove-Item Env:\PGPASSWORD
