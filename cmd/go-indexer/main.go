package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dovbysh/duf.git/internal/config"
	"github.com/dovbysh/duf.git/internal/database"
	"github.com/dovbysh/duf.git/internal/hasher"
	"github.com/dovbysh/duf.git/internal/models"
	"github.com/dovbysh/duf.git/internal/scanner"
)

type runOptions struct {
	configPath    string
	mode          string
	hashBatchSize int32
}

func main() {
	opts := parseOptions()

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		log.Fatal("Config error:", err)
	}

	db, err := database.NewClient(cfg.Database.DSN, cfg.Database.TableName)
	if err != nil {
		log.Fatal("DB error:", err)
	}
	defer db.Close()

	ctx := context.Background()
	err = db.Migrate()
	if err != nil {
		log.Fatal("DB migration error:", err)
	}

	if shouldRunScan(opts.mode) {
		scanFiles(ctx, db, cfg)
	}

	if shouldRunHash(opts.mode) {
		log.Printf("running processHashes...")
		for processHashes(ctx, db, cfg, opts.hashBatchSize) > 0 {
		}
	}
}

func parseOptions() runOptions {
	var opts runOptions
	var hashBatchSize int

	flag.StringVar(&opts.configPath, "config", "config.yaml", "path to config file")
	flag.StringVar(&opts.mode, "mode", "all", "run mode: all, scan, hash")
	flag.IntVar(&hashBatchSize, "hash-batch-size", 100, "number of files to hash per database query")
	flag.Parse()

	switch opts.mode {
	case "all", "scan", "hash":
	default:
		log.Fatalf("invalid mode %q, expected all, scan, or hash", opts.mode)
	}

	if hashBatchSize <= 0 {
		log.Fatalf("hash-batch-size must be greater than zero")
	}
	opts.hashBatchSize = int32(hashBatchSize)

	return opts
}

func shouldRunScan(mode string) bool {
	return mode == "all" || mode == "scan"
}

func shouldRunHash(mode string) bool {
	return mode == "all" || mode == "hash"
}

func scanFiles(ctx context.Context, db *database.PostgresClient, cfg *config.Config) {
	fileChan := make(chan models.FileRecord, 100)
	scanErrChan := make(chan error, 1)

	go func() {
		defer close(fileChan)

		s := scanner.NewScanner(cfg.Storage.ExcludePatterns)
		for _, path := range cfg.Storage.ScanPaths {
			if err := s.Scan(path, fileChan); err != nil {
				scanErrChan <- fmt.Errorf("scan %q: %w", path, err)
				return
			}
		}

		scanErrChan <- nil
	}()

	// Пакетная вставка в БД
	var batch []models.FileRecord
	for f := range fileChan {
		batch = append(batch, f)
		if len(batch) >= cfg.Database.BatchSize {
			if err := db.BatchReplace(ctx, batch); err != nil {
				log.Fatalf("BatchReplace error: %v", err)
			}
			batch = batch[:0]
		}
	}
	if err := db.BatchReplace(ctx, batch); err != nil {
		log.Fatalf("BatchReplace tail error: %v", err)
	}

	if err := <-scanErrChan; err != nil {
		log.Fatalf("Scan error: %v", err)
	}
}

func processHashes(ctx context.Context, db *database.PostgresClient, cfg *config.Config, batchSize int32) int64 {
	files, err := db.GetFilesWithoutHash(ctx, batchSize) // берем порцию
	if err != nil {
		log.Fatalf("processHashes GetFilesWithoutHash: %v", err)
	}
	if len(files) == 0 {
		log.Printf("No hashes found in %d files", batchSize)
		return 0
	}
	log.Printf("Found %d files,  batchSize is %d", len(files), batchSize)
	hashed := atomic.Int64{}
	var wg sync.WaitGroup
	jobs := make(chan models.FileRecord, len(files))

	// Воркеры
	for w := 1; w <= cfg.Performance.HashWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var (
				err  error
				hash string
			)
			for f := range jobs {
				hash, err = hasher.CalculateSHA256(f.Path)
				if err != nil {
					log.Printf("Error updating hash for file %v: %v", f.Path, err)
					continue
				}
				for i := 1; i <= 10; i++ {
					err = db.UpdateHash(ctx, f.ID, hash)
					if err != nil {
						sleeping := i * i
						log.Printf("Error updating hash for file %v: %v sleeping: %d", f.Path, err, sleeping)
						time.Sleep(time.Duration(sleeping) * time.Second)
						continue
					}
					break
				}
				if err != nil {
					log.Printf("Error updating hash for file %v: %v", f.Path, err)
				}

				fmt.Printf("Hashed: %s\n", f.Path)
				hashed.Add(1)
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)
	wg.Wait()

	return hashed.Load()
}
