package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dovbysh/duf.git/internal/ai/is_document"
	"github.com/dovbysh/duf.git/internal/ai/lmstudio"
	"github.com/dovbysh/duf.git/internal/config"
	"github.com/dovbysh/duf.git/internal/database"
	"github.com/dovbysh/duf.git/internal/hasher"
	"github.com/dovbysh/duf.git/internal/models"
	"github.com/dovbysh/duf.git/internal/scanner"
)

type runOptions struct {
	configPath          string
	mode                string
	hashBatchSize       int32
	migrateDownOne      bool
	applyDirtyMigration bool
}

func main() {
	opts := parseOptions()

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		log.Fatal("Config error:", err)
	}

	pool, err := cfg.DatabasePoolConfig()
	if err != nil {
		log.Fatal("DB pool config error:", err)
	}

	db, err := database.NewClient(cfg.Database.DSN, database.PoolConfig{
		MaxOpenConns:    pool.MaxOpenConns,
		MaxIdleConns:    pool.MaxIdleConns,
		ConnMaxLifetime: pool.ConnMaxLifetime,
		ConnMaxIdleTime: pool.ConnMaxIdleTime,
	})
	if err != nil {
		log.Fatal("DB error:", err)
	}
	defer db.Close()

	ctx := context.Background()
	if opts.migrateDownOne {
		if err := db.MigrateDownOne(); err != nil {
			log.Fatal("DB migrate down one error:", err)
		}
		log.Print("rolled back one database migration")
		return
	}

	if opts.applyDirtyMigration {
		applied, err := db.ApplyDirtyMigration()
		if err != nil {
			log.Fatal("DB apply dirty migration error:", err)
		}
		if applied {
			log.Print("applied current dirty database migration")
		} else {
			log.Print("no dirty database migration to apply")
		}
		return
	}

	err = db.Migrate()
	if err != nil {
		log.Fatal("DB migration error:", err)
	}

	if shouldRunScan(opts.mode) {
		scanFiles(ctx, db, cfg)
	}

	if shouldRunHash(opts.mode) {
		log.Printf("running processHashes...")
		lmClient := lmstudio.NewFromConfig(lmstudio.Config{
			AuthToken: cfg.LMStudio.AuthToken,
			APIURL:    cfg.LMStudio.APIURL,
			ModelName: cfg.LMStudio.ModelName,
		})
		for processHashes(ctx, db, cfg, lmClient, opts.hashBatchSize) > 0 {
		}
	}
}

func parseOptions() runOptions {
	var opts runOptions
	var hashBatchSize int

	flag.StringVar(&opts.configPath, "config", "config.yaml", "path to config file")
	flag.StringVar(&opts.mode, "mode", "all", "run mode: all, scan, hash")
	flag.IntVar(&hashBatchSize, "hash-batch-size", 100, "number of files to hash per database query")
	flag.BoolVar(&opts.migrateDownOne, "migrate-down-one", false, "roll back one database migration and exit")
	flag.BoolVar(&opts.applyDirtyMigration, "migrate-apply-dirty", false, "apply the current dirty database migration and exit")
	flag.Parse()

	if opts.migrateDownOne && opts.applyDirtyMigration {
		log.Fatal("migrate-down-one and migrate-apply-dirty cannot be used together")
	}

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

func processHashes(ctx context.Context, db *database.PostgresClient, cfg *config.Config, lmClient *lmstudio.Client, batchSize int32) int64 {
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
					continue
				}
				hashed.Add(1)
				fmt.Printf("Hashed: %s\n", f.Path)

				classifyDocumentImage(ctx, db, lmClient, f)
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

func classifyDocumentImage(ctx context.Context, db *database.PostgresClient, lmClient *lmstudio.Client, f models.FileRecord) {
	if !isSupportedDocumentImage(f.Path) {
		return
	}

	img, err := os.ReadFile(f.Path)
	if err != nil {
		log.Printf("Error reading image for document classification %v: %v", f.Path, err)
		return
	}

	message, err := lmClient.GetMessage(ctx, string(is_document.Prompt01), img)
	if err != nil {
		log.Printf("Error classifying document image %v: %v", f.Path, err)
		return
	}

	classification, err := is_document.GetDocumentClassification(message)
	if err != nil {
		log.Printf("Error parsing document classification for %v: %v", f.Path, err)
		return
	}

	if err := db.UpsertDocumentClassification(ctx, f.ID, *classification); err != nil {
		log.Printf("Error saving document classification for %v: %v", f.Path, err)
		return
	}
}

func isSupportedDocumentImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}
