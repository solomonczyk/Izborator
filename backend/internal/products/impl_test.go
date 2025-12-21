package products

import (
	"context"
	"testing"
	"time"

	"github.com/solomonczyk/izborator/internal/logger"
)

// mockStorage мок для Storage интерфейса
type mockStorage struct {
	searchProductsFunc    func(query string, limit, offset int) ([]*Product, int, error)
	getProductByIDFunc    func(id string) (*Product, error)
	browseProductsFunc    func(params BrowseParams) (*BrowseResult, error)
	saveProductFunc       func(product *Product) error
	//nolint:unused // может использоваться в будущем
	savePriceFunc         func(productID string, price float64, currency string) error
}

func (m *mockStorage) SearchProducts(query string, limit, offset int) ([]*Product, int, error) {
	if m.searchProductsFunc != nil {
		return m.searchProductsFunc(query, limit, offset)
	}
	return []*Product{}, 0, nil
}

func (m *mockStorage) GetProduct(id string) (*Product, error) {
	if m.getProductByIDFunc != nil {
		return m.getProductByIDFunc(id)
	}
	return nil, nil
}

func (m *mockStorage) Browse(ctx context.Context, params BrowseParams) (*BrowseResult, error) {
	if m.browseProductsFunc != nil {
		return m.browseProductsFunc(params)
	}
	return &BrowseResult{
		Items:      []BrowseProduct{},
		Total:      0,
		Page:       1,
		PerPage:    10,
		TotalPages: 0,
	}, nil
}

func (m *mockStorage) SaveProduct(product *Product) error {
	if m.saveProductFunc != nil {
		return m.saveProductFunc(product)
	}
	return nil
}

func (m *mockStorage) GetProductPrices(productID string) ([]*ProductPrice, error) {
	return nil, nil
}

func (m *mockStorage) SaveProductPrice(price *ProductPrice) error {
	return nil
}

func (m *mockStorage) GetURLsForRescrape(ctx context.Context, olderThan time.Duration, limit int) ([]RescrapeItem, error) {
	return nil, nil
}

// mockLogger мок для logger - используем реальный logger
func createMockLogger() *logger.Logger {
	return logger.New("info")
}

// TestSearchEmptyQuery тестирует Search с пустым запросом
func TestSearchEmptyQuery(t *testing.T) {
	service := &Service{
		storage: &mockStorage{},
		logger:  createMockLogger(),
	}

	ctx := context.Background()
	_, err := service.Search(ctx, "")

	if err != ErrInvalidSearchQuery {
		t.Errorf("expected ErrInvalidSearchQuery, got %v", err)
	}
}

// TestSearchValidQuery тестирует Search с валидным запросом
func TestSearchValidQuery(t *testing.T) {
	expectedProducts := []*Product{
		{ID: "1", Name: "iPhone 15"},
		{ID: "2", Name: "Samsung Galaxy S24"},
	}

	service := &Service{
		storage: &mockStorage{
			searchProductsFunc: func(query string, limit, offset int) ([]*Product, int, error) {
				if query != "phone" {
					t.Errorf("expected query 'phone', got %q", query)
				}
				if limit != 20 {
					t.Errorf("expected limit 20, got %d", limit)
				}
				if offset != 0 {
					t.Errorf("expected offset 0, got %d", offset)
				}
				return expectedProducts, 2, nil
			},
		},
		logger: createMockLogger(),
	}

	ctx := context.Background()
	products, err := service.Search(ctx, "phone")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(products) != 2 {
		t.Errorf("expected 2 products, got %d", len(products))
	}

	if products[0].ID != "1" {
		t.Errorf("expected product ID '1', got %q", products[0].ID)
	}
}

// TestSearchWithPaginationEmptyQuery тестирует SearchWithPagination с пустым запросом
func TestSearchWithPaginationEmptyQuery(t *testing.T) {
	service := &Service{
		storage: &mockStorage{},
		logger:  createMockLogger(),
	}

	ctx := context.Background()
	_, err := service.SearchWithPagination(ctx, "", 10, 0)

	if err != ErrInvalidSearchQuery {
		t.Errorf("expected ErrInvalidSearchQuery, got %v", err)
	}
}

// TestSearchWithPaginationValidQuery тестирует SearchWithPagination с валидным запросом
func TestSearchWithPaginationValidQuery(t *testing.T) {
	expectedProducts := []*Product{
		{ID: "1", Name: "iPhone 15"},
	}

	service := &Service{
		storage: &mockStorage{
			searchProductsFunc: func(query string, limit, offset int) ([]*Product, int, error) {
				return expectedProducts, 1, nil
			},
		},
		logger: createMockLogger(),
	}

	ctx := context.Background()
	result, err := service.SearchWithPagination(ctx, "phone", 10, 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 product, got %d", len(result.Items))
	}

	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

// TestSearchWithPaginationLimitBoundaries тестирует границы limit
func TestSearchWithPaginationLimitBoundaries(t *testing.T) {
	service := &Service{
		storage: &mockStorage{
			searchProductsFunc: func(query string, limit, offset int) ([]*Product, int, error) {
				// Проверяем, что limit корректно ограничен
				if limit < 1 {
					t.Errorf("limit should be at least 1, got %d", limit)
				}
				if limit > 100 {
					t.Errorf("limit should be at most 100, got %d", limit)
				}
				return []*Product{}, 0, nil
			},
		},
		logger: createMockLogger(),
	}

	ctx := context.Background()

	// Тест с limit = 0 (должен стать 20)
	_, err := service.SearchWithPagination(ctx, "test", 0, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Тест с limit > 100 (должен стать 100)
	_, err = service.SearchWithPagination(ctx, "test", 200, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Тест с отрицательным offset (должен стать 0)
	_, err = service.SearchWithPagination(ctx, "test", 10, -5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestGetByIDValidID тестирует GetByID с валидным ID
func TestGetByIDValidID(t *testing.T) {
	expectedProduct := &Product{
		ID:   "test-id",
		Name: "Test Product",
	}

	service := &Service{
		storage: &mockStorage{
			getProductByIDFunc: func(id string) (*Product, error) {
				if id != "test-id" {
					t.Errorf("expected id 'test-id', got %q", id)
				}
				return expectedProduct, nil
			},
		},
		logger: createMockLogger(),
	}

	product, err := service.GetByID("test-id")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if product == nil {
		t.Fatal("expected non-nil product")
	}

	if product.ID != "test-id" {
		t.Errorf("expected product ID 'test-id', got %q", product.ID)
	}
}

// TestGetByIDNotFound тестирует GetByID с несуществующим ID
func TestGetByIDNotFound(t *testing.T) {
	service := &Service{
		storage: &mockStorage{
			getProductByIDFunc: func(id string) (*Product, error) {
				return nil, ErrProductNotFound
			},
		},
		logger: createMockLogger(),
	}

	product, err := service.GetByID("non-existent")

	if err != ErrProductNotFound {
		t.Errorf("expected ErrProductNotFound, got %v", err)
	}

	if product != nil {
		t.Errorf("expected nil product, got %v", product)
	}
}

// TestBrowseValidParams тестирует Browse с валидными параметрами
func TestBrowseValidParams(t *testing.T) {

	service := &Service{
		storage: &mockStorage{
			browseProductsFunc: func(params BrowseParams) (*BrowseResult, error) {
				if params.Page < 1 {
					t.Errorf("page should be at least 1, got %d", params.Page)
				}
				if params.PerPage < 1 {
					t.Errorf("perPage should be at least 1, got %d", params.PerPage)
				}
				browseProducts := []BrowseProduct{
					{ID: "1", Name: "Product 1"},
					{ID: "2", Name: "Product 2"},
				}
				return &BrowseResult{
					Items:      browseProducts,
					Total:      2,
					Page:       params.Page,
					PerPage:    params.PerPage,
					TotalPages: 1,
				}, nil
			},
		},
		logger: createMockLogger(),
	}

	ctx := context.Background()
	params := BrowseParams{
		Page:    1,
		PerPage: 10,
	}

	result, err := service.Browse(ctx, params)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Items))
	}

	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
}

// TestBrowseDefaultParams тестирует Browse с дефолтными параметрами
func TestBrowseDefaultParams(t *testing.T) {
	service := &Service{
		storage: &mockStorage{
			browseProductsFunc: func(params BrowseParams) (*BrowseResult, error) {
				// Проверяем, что дефолтные значения применены
				if params.Page < 1 {
					t.Errorf("page should be at least 1, got %d", params.Page)
				}
				if params.PerPage < 1 {
					t.Errorf("perPage should be at least 1, got %d", params.PerPage)
				}
				return &BrowseResult{
					Items:      []BrowseProduct{},
					Total:      0,
					Page:       params.Page,
					PerPage:    params.PerPage,
					TotalPages: 0,
				}, nil
			},
		},
		logger: createMockLogger(),
	}

	ctx := context.Background()
	params := BrowseParams{} // Пустые параметры

	result, err := service.Browse(ctx, params)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

