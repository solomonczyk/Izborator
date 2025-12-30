package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/logger"
)

// BaseAdapter базовый адаптер с общей функциональностью для всех хранилищ
type BaseAdapter struct {
	pg     *Postgres
	ctx    context.Context
	logger *logger.Logger
}

// NewBaseAdapter создает новый базовый адаптер
func NewBaseAdapter(pg *Postgres, logger *logger.Logger) *BaseAdapter {
	return &BaseAdapter{
		pg:     pg,
		ctx:    pg.Context(),
		logger: logger,
	}
}

// HandleQueryError обрабатывает ошибки запроса к базе данных
// Преобразует pgx ошибки в понятные domain ошибки
func (a *BaseAdapter) HandleQueryError(operation string, err error) error {
	if err == nil {
		return nil
	}

	if err == pgx.ErrNoRows {
		if a.logger != nil {
			a.logger.Warn("No rows found", map[string]interface{}{
				"operation": operation,
			})
		}
		return fmt.Errorf("not found")
	}

	if a.logger != nil {
		a.logger.Error("Database query error", map[string]interface{}{
			"operation": operation,
			"error":     err.Error(),
		})
	}

	return fmt.Errorf("database error: %w", err)
}

// ParseUUID парсит строку в UUID с проверкой
func (a *BaseAdapter) ParseUUID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		if a.logger != nil {
			a.logger.Warn("Invalid UUID format", map[string]interface{}{
				"id":    id,
				"error": err.Error(),
			})
		}
		return uuid.UUID{}, fmt.Errorf("invalid ID format: %w", err)
	}
	return parsed, nil
}

// LogQuery логирует операцию с базой данных
func (a *BaseAdapter) LogQuery(operation string, details map[string]interface{}) {
	if a.logger == nil {
		return
	}
	if details == nil {
		details = make(map[string]interface{})
	}
	details["operation"] = operation

	a.logger.Info("Database operation", details)
}

// LogError логирует ошибку операции с базой данных
func (a *BaseAdapter) LogError(operation string, err error, details map[string]interface{}) {
	if a.logger == nil {
		return
	}
	if details == nil {
		details = make(map[string]interface{})
	}
	details["operation"] = operation
	if err != nil {
		details["error"] = err.Error()
	}

	a.logger.Error("Database operation failed", details)
}

// GetContext возвращает контекст адаптера
func (a *BaseAdapter) GetContext() context.Context {
	return a.ctx
}

// GetLogger возвращает логгер адаптера
func (a *BaseAdapter) GetLogger() *logger.Logger {
	return a.logger
}

// GetPostgres возвращает экземпляр Postgres для доступа к БД
func (a *BaseAdapter) GetPostgres() *Postgres {
	return a.pg
}
