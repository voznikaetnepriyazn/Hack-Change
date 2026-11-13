package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/httpclient"
	chunker "github.com/voznikaetnepriyazn/Hack-Change/agent/rag/chunk"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/rag/embed"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/rag/pipeline"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/rag/storage"
)

const (
	defaultDataDir      = "agent/tools/data"
	defaultChunkSize    = 1000
	defaultChunkOverlap = 150
)

type documentPipeline interface {
	IngestDocument(context.Context, chunker.Document) error
}

func main() {
	if err := run(); err != nil {
		slog.Error("RAG ingestion failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load environment file: %w", err)
	}

	dataDir := flag.String("data-dir", defaultDataDir, "directory containing Markdown documents")
	chunkSize := flag.Int("chunk-size", defaultChunkSize, "maximum chunk size in Unicode code points")
	chunkOverlap := flag.Int("chunk-overlap", defaultChunkOverlap, "chunk overlap in Unicode code points")
	flag.Parse()

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	apiKey := strings.TrimSpace(os.Getenv("EMBED_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("EMBED_KEY"))
	}
	if apiKey == "" {
		return errors.New("EMBED_KEY is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("configure PostgreSQL connection: %w", err)
	}
	defer pool.Close()

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	err = pool.Ping(connectCtx)
	cancel()
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	docChunker, err := chunker.NewRecursiveChunker(*chunkSize, *chunkOverlap)
	if err != nil {
		return fmt.Errorf("configure document chunker: %w", err)
	}
	var embeddingHTTPClient *http.Client
	httpsProxy := strings.TrimSpace(os.Getenv("HTTPS_PROXY"))
	httpProxy := strings.TrimSpace(os.Getenv("HTTP_PROXY"))
	if httpsProxy != "" || httpProxy != "" {
		embeddingHTTPClient, err = httpclient.NewSOCKS5ProxyClient(httpsProxy, httpProxy)
		if err != nil {
			return fmt.Errorf("configure embeddings HTTP client: %w", err)
		}
	} else {
		embeddingHTTPClient = &http.Client{Timeout: 180 * time.Second}
	}
	embedder := embed.NewEmbedderWithHTTPClient(
		apiKey,
		os.Getenv("EMBED_URL"),
		os.Getenv("EMBED_MODEL"),
		embeddingHTTPClient,
	)
	ragPipeline := pipeline.NewPipeline(docChunker, embedder, storage.NewStore(pool))

	documentCount, err := ingestMarkdownDirectory(ctx, *dataDir, ragPipeline)
	if err != nil {
		return err
	}
	slog.Info("RAG documents ingested", "documents", documentCount, "directory", *dataDir)
	return nil
}

func ingestMarkdownDirectory(ctx context.Context, dataDir string, ingestor documentPipeline) (int, error) {
	if strings.TrimSpace(dataDir) == "" {
		return 0, errors.New("data directory is required")
	}
	if ingestor == nil {
		return 0, errors.New("document pipeline is required")
	}

	documentCount := 0
	err := filepath.WalkDir(dataDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("access %q: %w", path, walkErr)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() || !entry.Type().IsRegular() || !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read document %q: %w", path, err)
		}
		relativePath, err := filepath.Rel(dataDir, path)
		if err != nil {
			return fmt.Errorf("resolve document source %q: %w", path, err)
		}
		source := filepath.ToSlash(relativePath)
		metadata := map[string]any{"source": source}
		directory := filepath.ToSlash(filepath.Dir(relativePath))
		if directory != "." {
			metadata["category"] = strings.Split(directory, "/")[0]
		}

		if err := ingestor.IngestDocument(ctx, chunker.Document{
			Content:  string(content),
			Metadata: metadata,
		}); err != nil {
			return fmt.Errorf("ingest document %q: %w", source, err)
		}
		documentCount++
		slog.Info("RAG document ingested", "source", source)
		return nil
	})
	if err != nil {
		return documentCount, fmt.Errorf("scan Markdown documents in %q: %w", dataDir, err)
	}
	if documentCount == 0 {
		return 0, fmt.Errorf("no Markdown documents found in %q", dataDir)
	}
	return documentCount, nil
}
