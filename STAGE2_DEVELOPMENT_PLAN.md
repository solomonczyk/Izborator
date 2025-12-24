# Stage 2 Development Plan - Important Fixes

## Overview
**Duration:** 24 hours (5 days × ~5 hours)  
**Status:** 🚀 IN PROGRESS  
**Start Date:** December 24, 2025

---

## Day 1-2: Storage Adapter Analysis & Refactoring (~8 hours)

### Current State Analysis

#### Storage Adapters Found (12 total)
1. ProductsAdapter (1,245 lines)
2. CategoriesAdapter (253 lines)
3. CitiesAdapter
4. PriceHistoryAdapter
5. ScraperAdapter
6. ProcessorAdapter
7. MatchingAdapter
8. ClassifierAdapter
9. AutoconfigAdapter
10. AttributesAdapter
11. ProductTypesAdapter
12. ScrapingStatsAdapter

#### Common Patterns Identified
- All have `Postgres` and `context.Context` fields
- Some have `logger.Logger` and `Meilisearch` fields
- Repeated error handling patterns
- Similar query structure across adapters
- Duplicate logging patterns

#### Estimated Code Duplication
- ~200+ lines of duplicated CRUD patterns
- ~50+ lines of error handling boilerplate
- ~30+ lines of logging patterns

### Strategy

#### Phase 1: Create BaseAdapter (3 hours)
```go
// internal/storage/base_adapter.go

type BaseAdapter struct {
    pg     *Postgres
    ctx    context.Context
    logger *logger.Logger
}

// Common methods:
// - handleQueryError(err error) error
// - parseUUID(id string) (uuid.UUID, error)
// - logQuery(operation string, details map[string]interface{})
// - logError(operation string, err error)
// - scanRow(rows pgx.Rows, dest ...interface{}) error
```

#### Phase 2: Refactor Adapters (5 hours)
- ProductsAdapter - embed BaseAdapter
- CategoriesAdapter - embed BaseAdapter
- CitiesAdapter - embed BaseAdapter
- PriceHistoryAdapter - embed BaseAdapter
- Other adapters - update as needed

### Success Criteria
- [ ] All adapters compile
- [ ] No behavior changes
- [ ] ~150+ lines eliminated
- [ ] Tests pass
- [ ] git diff shows only refactoring

---

## Day 3: Error Handling Standardization (~6 hours)

### Current Issues
- Different error response formats across modules
- Inconsistent HTTP status codes
- Database errors not properly translated
- Validation errors mixed with business logic errors

### Plan
1. Create unified `ErrorResponse` structure
2. Add error translation layer
3. Standardize HTTP status codes
4. Document error codes

### Deliverables
- `internal/http/response/error_response.go`
- Updated all handlers to use consistent format
- Error handling guide in docs

---

## Day 4: Request Validation Framework (~6 hours)

### Current Issues
- Input validation scattered across handlers
- No consistent sanitization
- Business logic validation mixed with data validation
- Use of `validator/v10` package not consistent

### Plan
1. Create validation middleware
2. Implement request validators
3. Add input sanitization
4. Create validation error responses

### Deliverables
- `internal/http/middleware/validation.go`
- Validators for each major endpoint
- Validation testing

---

## Day 5: Logging & Monitoring (~4 hours)

### Current Status
- Zerolog configured
- Basic logging in place
- No structured logging patterns
- No metrics/monitoring

### Plan
1. Add structured logging patterns
2. Implement health check endpoints
3. Add basic metrics collection
4. Document logging best practices

### Deliverables
- Logging guidelines
- Health check endpoints
- Metrics dashboard setup

---

## Success Metrics

| Metric | Goal | Current |
|--------|------|---------|
| Code Duplication | 0% | ~5% |
| Test Coverage | 40%+ | 30% |
| API Consistency | 100% | ~70% |
| Error Handling | Standardized | Varied |
| Monitoring | Setup | None |

---

## Next Steps After Stage 2

If time permits, proceed to **Stage 3 (Nice-to-have improvements)**:
- Complete unit test coverage (60%+)
- E2E tests for frontend
- Performance optimization
- API documentation (OpenAPI/Swagger)

---

## Notes

- All changes will be committed with clear commit messages
- Each day's work will be validated before proceeding
- Documentation will be updated as we progress
- Git history will remain clean

