$env:PGPASSWORD = "yasdas230321*"
$psqlPath = "C:\Program Files\PostgreSQL\18\bin\psql.exe"

Write-Host "Checking uploaded documents and chunks..." -ForegroundColor Cyan

& $psqlPath -U postgres -d postgres -c "
-- Check uploaded document
SELECT id, filename, file_type, file_size_bytes, document_type, processing_status, created_at
FROM uploaded_documents
ORDER BY created_at DESC
LIMIT 1;

-- Check document chunks
SELECT id, document_id, chunk_index, chunk_text, token_count, created_at
FROM document_chunks
WHERE document_id = (SELECT id FROM uploaded_documents ORDER BY created_at DESC LIMIT 1)
ORDER BY chunk_index;

-- Count chunks
SELECT COUNT(*) as total_chunks
FROM document_chunks
WHERE document_id = (SELECT id FROM uploaded_documents ORDER BY created_at DESC LIMIT 1);
"

Remove-Item Env:\PGPASSWORD
