$env:PGPASSWORD = "yasdas230321*"
$psqlPath = "C:\Program Files\PostgreSQL\18\bin\psql.exe"

Write-Host "Checking document processing results..." -ForegroundColor Cyan

& $psqlPath -U postgres -d postgres -c "
-- Check uploaded document
SELECT 
    id, 
    filename, 
    file_size_bytes, 
    processing_status,
    CASE WHEN created_at IS NOT NULL THEN 'Created' ELSE 'No' END as has_timestamp
FROM uploaded_documents
ORDER BY created_at DESC
LIMIT 1;

-- Check document chunks
SELECT 
    id, 
    chunk_index, 
    LEFT(content, 80) as content_preview,
    token_count,
    CASE WHEN embedding IS NOT NULL THEN 'Yes' ELSE 'No' END as has_embedding
FROM document_chunks
WHERE document_id = (SELECT id FROM uploaded_documents ORDER BY created_at DESC LIMIT 1)
ORDER BY chunk_index
LIMIT 5;

-- Count total chunks
SELECT COUNT(*) as total_chunks
FROM document_chunks
WHERE document_id = (SELECT id FROM uploaded_documents ORDER BY created_at DESC LIMIT 1);
"

Write-Host ""
Write-Host "Document processing verification complete!" -ForegroundColor Green

Remove-Item Env:\PGPASSWORD
