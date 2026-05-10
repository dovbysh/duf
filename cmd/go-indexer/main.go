package main

import (
	"context"
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
	//err = db.InitSchema(ctx)
	//if err != nil {
	//	log.Fatal("DB InitSchema error:", err)
	//}
	//
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
	for processHashes(ctx, db, cfg) > 0 {

	}
}

func processHashes(ctx context.Context, db *database.ManticoreClient, cfg *config.Config) int64 {
	files, err := db.GetFilesWithoutHash(ctx, 100) // берем порцию
	if err != nil {
		log.Fatalf("processHashes GetFilesWithoutHash: %v", err)
	}
	if len(files) == 0 {
		return 0
	}

	hashed := atomic.Int64{}
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
					var herror error
					for i := 1; i <= 10; i++ {
						herror = db.UpdateHash(ctx, f.ID, hash)
						if herror != nil {
							sleeping := i * i
							log.Printf("Error updating hash for file %v: %v sleeping: %d", f.Path, herror, sleeping)
							time.Sleep(time.Duration(sleeping) * time.Second)
							continue
						}
						break
					}
					if herror != nil {
						log.Printf("Error updating hash for file %v: %v", f.Path, herror)
					}

					fmt.Printf("Hashed: %s\n", f.Path)
					hashed.Add(1)
				}
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
