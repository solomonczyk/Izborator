package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/storage"
)

func main() {
	var (
		reindex = flag.Bool("reindex", false, "Reindex all products (clear and rebuild index)")
		sync    = flag.Bool("sync", false, "Sync products from PostgreSQL to Meilisearch")
		setup   = flag.Bool("setup", false, "Setup Meilisearch index (configure searchable fields, filters)")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logger.New(cfg.LogLevel)

	// Инициализация storage
	pg, err := storage.NewPostgres(&cfg.DB, logger)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pg.Close()

	meili, err := storage.NewMeilisearch(&cfg.Meili, logger)
	if err != nil {
		log.Fatalf("Failed to connect to Meilisearch: %v", err)
	}

	indexer := NewIndexer(pg, meili, logger)

	switch {
	case *setup:
		if err := indexer.SetupIndex(); err != nil {
			log.Fatalf("Failed to setup index: %v", err)
		}
		fmt.Println("Index setup completed successfully")
	case *reindex:
		if err := indexer.ReindexAll(); err != nil {
			log.Fatalf("Failed to reindex: %v", err)
		}
		fmt.Println("Reindexing completed successfully")
	case *sync:
		if err := indexer.SyncProducts(); err != nil {
			log.Fatalf("Failed to sync: %v", err)
		}
		fmt.Println("Sync completed successfully")
	default:
		flag.Usage()
		log.Fatal("Please specify an action: -setup, -reindex, or -sync")
	}
}

// Indexer индексирует товары в Meilisearch
type Indexer struct {
	pg     *storage.Postgres
	meili  *storage.Meilisearch
	logger *logger.Logger
}

// NewIndexer создаёт новый индексатор
func NewIndexer(pg *storage.Postgres, meili *storage.Meilisearch, log *logger.Logger) *Indexer {
	return &Indexer{
		pg:     pg,
		meili:  meili,
		logger: log,
	}
}

// SetupIndex настраивает индекс Meilisearch
func (i *Indexer) SetupIndex() error {
	index := i.meili.Client().Index("products")

	// Настройка поисковых полей
	searchableAttributes := []string{
		"name",
		"description",
		"brand",
		"category",
	}

	// Настройка фильтруемых полей
	filterableAttributes := []string{
		"brand",
		"category",
		"category_id",
		"created_at",
		"updated_at",
	}

	// Настройка сортируемых полей
	sortableAttributes := []string{
		"name",
		"brand",
		"category",
		"created_at",
		"updated_at",
	}

	// Настройка индекса
	settings := &meilisearch.Settings{
		SearchableAttributes: searchableAttributes,
		FilterableAttributes: filterableAttributes,
		SortableAttributes:   sortableAttributes,
		RankingRules: []string{
			"words",
			"typo",
			"proximity",
			"attribute",
			"sort",
			"exactness",
		},
		StopWords: []string{}, // Можно добавить стоп-слова для сербского языка
		Synonyms:  map[string][]string{},
	}

	_, err := index.UpdateSettings(settings)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	i.logger.Info("Index settings updated", map[string]interface{}{
		"searchable": searchableAttributes,
		"filterable": filterableAttributes,
		"sortable":   sortableAttributes,
	})

	return nil
}

// ReindexAll переиндексирует все товары
func (i *Indexer) ReindexAll() error {
	index := i.meili.Client().Index("products")

	// Удаляем все документы
	if _, err := index.DeleteAllDocuments(); err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	i.logger.Info("Deleted all documents from index")

	// Синхронизируем все товары
	return i.SyncProducts()
}

// SyncProducts синхронизирует товары из PostgreSQL в Meilisearch
func (i *Indexer) SyncProducts() error {
	index := i.meili.Client().Index("products")

	// Получаем все товары из PostgreSQL
	query := `
		SELECT id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
	`

	ctx := context.Background()
	rows, err := i.pg.DB().Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var documents []map[string]interface{}
	batchSize := 1000
	total := 0

	for rows.Next() {
		var (
			id          string
			name        string
			description *string
			brand       *string
			category    *string
			categoryID  *string
			imageURL    *string
			specsJSON   []byte
			createdAt   time.Time
			updatedAt   time.Time
		)

		if err := rows.Scan(
			&id,
			&name,
			&description,
			&brand,
			&category,
			&categoryID,
			&imageURL,
			&specsJSON,
			&createdAt,
			&updatedAt,
		); err != nil {
			return fmt.Errorf("failed to scan product: %w", err)
		}

		// Создаём документ для Meilisearch
		doc := map[string]interface{}{
			"id":         id,
			"name":       name,
			"created_at": createdAt.Format(time.RFC3339),
			"updated_at": updatedAt.Format(time.RFC3339),
		}

		if description != nil {
			doc["description"] = *description
		}
		if brand != nil {
			doc["brand"] = *brand
		}
		if category != nil {
			doc["category"] = *category
		}
		if categoryID != nil {
			doc["category_id"] = *categoryID
		}
		if imageURL != nil {
			doc["image_url"] = *imageURL
		}

		// Парсим specs из JSONB
		if len(specsJSON) > 0 {
			var specs map[string]interface{}
			if err := json.Unmarshal(specsJSON, &specs); err == nil {
				doc["specs"] = specs
			}
		}

		// Получаем названия магазинов для этого товара
		shopNamesQuery := `
			SELECT DISTINCT s.name 
			FROM product_prices pp
			JOIN shops s ON pp.shop_id = s.id
			WHERE pp.product_id = $1
		`
		shopRows, err := i.pg.DB().Query(ctx, shopNamesQuery, id)
		if err == nil {
			var shopNames []string
			for shopRows.Next() {
				var shopName string
				if err := shopRows.Scan(&shopName); err == nil && shopName != "" {
					shopNames = append(shopNames, shopName)
				}
			}
			shopRows.Close()
			
			if len(shopNames) > 0 {
				doc["shop_names"] = shopNames
				doc["shops_count"] = len(shopNames)
			}
		}

		documents = append(documents, doc)

		// Отправляем батчами
		if len(documents) >= batchSize {
			if err := i.indexBatch(index, documents); err != nil {
				return err
			}
			total += len(documents)
			i.logger.Info("Indexed batch", map[string]interface{}{
				"count": len(documents),
				"total": total,
			})
			documents = documents[:0]
		}
	}

	// Отправляем оставшиеся документы
	if len(documents) > 0 {
		if err := i.indexBatch(index, documents); err != nil {
			return err
		}
		total += len(documents)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating products: %w", err)
	}

	i.logger.Info("Sync completed", map[string]interface{}{
		"total_products": total,
	})

	return nil
}

// indexBatch индексирует батч документов
func (i *Indexer) indexBatch(index *meilisearch.Index, documents []map[string]interface{}) error {
	task, err := index.AddDocuments(documents, "id")
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}

	// Ждём завершения задачи
	for {
		taskInfo, err := index.GetTask(task.TaskUID)
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}

		if taskInfo.Status == "succeeded" {
			break
		}
		if taskInfo.Status == "failed" {
			return fmt.Errorf("indexing task failed: %v", taskInfo.Error)
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
