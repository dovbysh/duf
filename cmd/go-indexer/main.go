package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/dovbysh/duf.git/internal/config"
	"github.com/dovbysh/duf.git/internal/database"
	"github.com/dovbysh/duf.git/internal/hasher"
	"github.com/dovbysh/duf.git/internal/models"
	"github.com/dovbysh/duf.git/internal/scanner"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal("Config error:", err)
	}

	db, err := database.NewClient(cfg.Database.DSN, cfg.Database.TableName)
	if err != nil {
		log.Fatal("DB error:", err)
	}

	ctx := context.Background()
	err = db.InitSchema(ctx)
	if err != nil {
		log.Fatal("DB InitSchema error:", err)
	}

	fileChan := make(chan models.FileRecord, 100)
	go func() {
		s := scanner.NewScanner(cfg.Storage.ExcludePatterns)
		for _, path := range cfg.Storage.ScanPaths {
			s.Scan(path, fileChan)
		}
		close(fileChan)
	}()

	// Пакетная вставка в БД
	var batch []models.FileRecord
	for f := range fileChan {
		batch = append(batch, f)
		if len(batch) >= cfg.Database.BatchSize {
			db.BatchReplace(ctx, batch)
			batch = batch[:0]
		}
	}
	db.BatchReplace(ctx, batch) // хвост

	// 3. Расчет Хешей
	processHashes(ctx, db, cfg)
}

func processHashes(ctx context.Context, db *database.ManticoreClient, cfg *config.Config) {
	files, _ := db.GetFilesWithoutHash(ctx, 5000) // берем порцию
	if len(files) == 0 {
		return
	}

	var wg sync.WaitGroup
	jobs := make(chan models.FileRecord, len(files))

	// Воркеры
	for w := 1; w <= cfg.Performance.HashWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				hash, err := hasher.CalculateSHA256(f.Path)
				if err == nil {
					db.UpdateHash(ctx, f.ID, hash)
					fmt.Printf("Hashed: %s\n", f.Path)
				}
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)
	wg.Wait()
}
