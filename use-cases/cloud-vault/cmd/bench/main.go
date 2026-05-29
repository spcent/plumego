package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"cloud-vault/internal/config"
	"cloud-vault/internal/database"
	"cloud-vault/internal/document"
	"cloud-vault/internal/search"
	"cloud-vault/internal/storage"
)

func main() {
	var (
		docCount      = flag.Int("docs", 100, "Number of documents to create")
		searchQueries = flag.Int("queries", 10, "Number of search queries to run")
		outputJSON    = flag.Bool("json", false, "Output results as JSON")
	)
	flag.Parse()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "cloud-vault-bench-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "bench.db")
	storagePath := filepath.Join(tmpDir, "storage")

	// Initialize database
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize storage
	store := storage.NewLocalStorage(storagePath)

	// Initialize document service
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	// Initialize search service
	searchRepo := search.NewRepository(db)
	searchEngine := search.NewFTSEngine(db, 50)
	searchSvc := search.NewService(searchEngine, searchRepo, store, config.SearchConfig{
		Enabled:     true,
		IndexOnSave: true,
	})

	// Wire index hook
	docSvc.SetIndexHook(func(ctx context.Context, ev document.IndexEvent) {
		searchSvc.HandleIndexEvent(ctx, search.IndexEvent{
			DocID:    ev.DocID,
			Content:  ev.Content,
			Version:  ev.Version,
			Hash:     ev.Hash,
			Deleted:  ev.Deleted,
			IsImport: ev.IsImport,
		})
	})

	ctx := context.Background()

	// Benchmark: Create documents
	startCreate := time.Now()
	var docIDs []string
	for i := 0; i < *docCount; i++ {
		title := fmt.Sprintf("Benchmark Document %d", i)
		content := fmt.Sprintf("# Benchmark Document %d\n\nThis is document number %d for performance testing.\n\nIt contains multiple paragraphs to simulate real content.\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit.", i, i)

		result, err := docSvc.Create(ctx, document.CreateRequest{
			Title:   title,
			Content: content,
		})
		if err != nil {
			log.Fatalf("Failed to create document %d: %v", i, err)
		}
		docIDs = append(docIDs, result.ID)
	}
	durationCreate := time.Since(startCreate)

	// Benchmark: Search queries
	searchTerms := []string{"benchmark", "document", "performance", "testing", "lorem"}
	startSearch := time.Now()
	var totalSearchDuration time.Duration
	for i := 0; i < *searchQueries; i++ {
		term := searchTerms[i%len(searchTerms)]
		queryStart := time.Now()
		_, err := searchSvc.Search(ctx, search.SearchQuery{
			Q:     term,
			Limit: 10,
		})
		if err != nil {
			log.Fatalf("Search query %d failed: %v", i, err)
		}
		totalSearchDuration += time.Since(queryStart)
	}
	durationSearch := time.Since(startSearch)
	avgSearchTime := totalSearchDuration / time.Duration(*searchQueries)

	// Get index status
	indexStatus, err := searchSvc.GetIndexStatus(ctx)
	if err != nil {
		log.Fatalf("Failed to get index status: %v", err)
	}

	// Output results
	if *outputJSON {
		fmt.Printf(`{"doc_count":%d,"create_duration_ms":%d,"search_queries":%d,"search_duration_ms":%d,"avg_search_ms":%d,"indexed_docs":%d}`,
			*docCount,
			durationCreate.Milliseconds(),
			*searchQueries,
			durationSearch.Milliseconds(),
			avgSearchTime.Milliseconds(),
			indexStatus.Indexed,
		)
	} else {
		fmt.Printf("Cloud Vault Benchmark Results\n")
		fmt.Printf("============================\n")
		fmt.Printf("Documents created: %d\n", *docCount)
		fmt.Printf("Create duration: %v (%.2f ms/doc)\n", durationCreate, float64(durationCreate.Milliseconds())/float64(*docCount))
		fmt.Printf("Search queries: %d\n", *searchQueries)
		fmt.Printf("Total search duration: %v\n", durationSearch)
		fmt.Printf("Average search time: %v\n", avgSearchTime)
		fmt.Printf("Indexed documents: %d\n", indexStatus.Indexed)
		fmt.Printf("Index coverage: %.1f%%\n", float64(indexStatus.Indexed)/float64(*docCount)*100)
	}
}
