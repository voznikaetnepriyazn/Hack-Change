DROP INDEX IF EXISTS chunks_embedding_hnsw_idx;

-- Embeddings from the previous model are incompatible and must be regenerated.
ALTER TABLE chunks
    DROP COLUMN embedding,
    ADD COLUMN embedding vector(1024);

CREATE INDEX chunks_embedding_hnsw_idx
    ON chunks USING hnsw (embedding vector_cosine_ops);
