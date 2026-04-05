-- name: CreateDocument :one
INSERT INTO documents (name, storage_key, file_type, file_hash, size_bytes, uploaded_by, status)
VALUES ($1, $2, $3, $4, $5, $6, 'pending')
RETURNING *;

-- name: GetDocument :one
SELECT * FROM documents WHERE id = $1;

-- name: GetDocumentByHash :one
SELECT * FROM documents WHERE file_hash = $1;

-- name: ListDocumentsByUser :many
SELECT * FROM documents WHERE uploaded_by = $1 ORDER BY created_at DESC LIMIT $2;

-- name: ListDocuments :many
SELECT * FROM documents ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateDocumentStatus :exec
UPDATE documents SET status = $1, updated_at = now() WHERE id = $2;

-- name: UpdateDocumentStatusWithError :exec
UPDATE documents SET status = $1, error = $2, updated_at = now() WHERE id = $3;

-- name: UpdateDocumentPages :exec
UPDATE documents SET total_pages = $1, updated_at = now() WHERE id = $2;

-- name: DeleteDocument :exec
DELETE FROM documents WHERE id = $1;

-- name: InsertDocumentPage :exec
INSERT INTO document_pages (document_id, page_number, text, tables, images)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (document_id, page_number) DO UPDATE SET
    text = EXCLUDED.text,
    tables = EXCLUDED.tables,
    images = EXCLUDED.images;

-- name: GetDocumentPages :many
SELECT * FROM document_pages
WHERE document_id = $1
ORDER BY page_number;

-- name: GetDocumentPageRange :many
SELECT * FROM document_pages
WHERE document_id = $1 AND page_number >= $2 AND page_number <= $3
ORDER BY page_number;

-- name: InsertDocumentTree :one
INSERT INTO document_trees (document_id, tree, doc_description, model_used, node_count)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetDocumentTree :one
SELECT * FROM document_trees
WHERE document_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: GetDocumentTreeByID :one
SELECT * FROM document_trees WHERE id = $1;

-- name: CreateCollection :one
INSERT INTO collections (name, description) VALUES ($1, $2) RETURNING *;

-- name: ListCollections :many
SELECT * FROM collections ORDER BY name;

-- name: AddDocumentToCollection :exec
INSERT INTO collection_documents (collection_id, document_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListCollectionDocuments :many
SELECT d.* FROM documents d
JOIN collection_documents cd ON cd.document_id = d.id
WHERE cd.collection_id = $1
ORDER BY d.created_at DESC;

-- name: GetAllDocumentTrees :many
SELECT dt.* FROM document_trees dt
JOIN documents d ON d.id = dt.document_id
WHERE d.status = 'ready'
ORDER BY dt.created_at DESC;

-- name: GetCollectionDocumentTrees :many
SELECT dt.* FROM document_trees dt
JOIN collection_documents cd ON cd.document_id = dt.document_id
WHERE cd.collection_id = $1
ORDER BY dt.created_at DESC;
