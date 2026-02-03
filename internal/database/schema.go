package database

// SQL schema for pgvector-based RAG storage
const SchemaSQL = `
-- Create extension if not exists
CREATE EXTENSION IF NOT EXISTS vector;

-- Chunks table for storing code/doc chunks with embeddings
CREATE TABLE IF NOT EXISTS chunks (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,

    -- Source information
    source VARCHAR(50) NOT NULL,  -- 'repo', 'docs', 'external'
    type VARCHAR(50) NOT NULL,    -- 'code', 'doc', 'config', 'test'
    path TEXT NOT NULL,
    language VARCHAR(50),
    symbol VARCHAR(255),          -- function/class name if applicable
    component VARCHAR(255),       -- module/component name

    -- Content
    content TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,  -- SHA256 of content for deduplication

    -- Embedding
    embedding vector(1536),  -- Dimension for text-embedding-ada-002 or similar
    embedding_model VARCHAR(100),  -- Model used to generate this embedding (e.g., "bge-code", "text-embedding-3-small")

    -- Metadata
    metadata JSONB,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    commit_sha VARCHAR(40),  -- Git commit SHA if available

    -- Indexes
    CONSTRAINT unique_chunk UNIQUE (project_id, content_hash)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_chunks_project_id ON chunks(project_id);
CREATE INDEX IF NOT EXISTS idx_chunks_type ON chunks(type);
CREATE INDEX IF NOT EXISTS idx_chunks_path ON chunks(path);
CREATE INDEX IF NOT EXISTS idx_chunks_source ON chunks(source);
CREATE INDEX IF NOT EXISTS idx_chunks_symbol ON chunks(symbol);
CREATE INDEX IF NOT EXISTS idx_chunks_commit ON chunks(commit_sha);
CREATE INDEX IF NOT EXISTS idx_chunks_metadata ON chunks USING gin(metadata);

-- Vector similarity index (using HNSW for better performance)
CREATE INDEX IF NOT EXISTS idx_chunks_embedding ON chunks USING hnsw (embedding vector_cosine_ops);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_chunks_updated_at ON chunks;
CREATE TRIGGER update_chunks_updated_at BEFORE UPDATE ON chunks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
`

// GetSchemaSQL returns the schema SQL
func GetSchemaSQL() string {
	return SchemaSQL
}
