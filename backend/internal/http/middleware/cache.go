package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/solomonczyk/izborator/internal/logger"
)

// CacheMiddleware создаёт middleware для кэширования HTTP ответов
func CacheMiddleware(redisClient *redis.Client, log *logger.Logger, ttl time.Duration) func(http.Handler) http.Handler {
	if redisClient == nil {
		// Если Redis недоступен, возвращаем пустой middleware
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Только GET запросы кэшируем
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Генерируем ключ кэша на основе URL и query параметров
			cacheKey := generateCacheKey(r)

			// Пытаемся получить из кэша
			ctx := r.Context()
			cached, err := redisClient.Get(ctx, cacheKey).Bytes()
			if err == nil {
				// Нашли в кэше - возвращаем
				var cachedResponse cachedResponse
				if err := cachedResponse.Unmarshal(cached); err == nil {
					// Устанавливаем заголовки
					for k, v := range cachedResponse.Headers {
						w.Header().Set(k, v)
					}
					w.Header().Set("X-Cache", "HIT")
					w.WriteHeader(cachedResponse.StatusCode)
					_, _ = w.Write(cachedResponse.Body)
					return
				}
			}

			// Не нашли в кэше - выполняем запрос и сохраняем результат
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
				headers:        make(http.Header),
			}

			next.ServeHTTP(recorder, r)

			// Устанавливаем заголовок X-Cache
			w.Header().Set("X-Cache", "MISS")

			// Сохраняем в кэш только успешные ответы (2xx)
			if recorder.statusCode >= 200 && recorder.statusCode < 300 {
				response := cachedResponse{
					StatusCode: recorder.statusCode,
					Headers:    make(map[string]string),
					Body:       recorder.body.Bytes(),
				}

				// Копируем заголовки (исключаем некоторые)
				for k, v := range recorder.headers {
					if k != "Content-Length" && k != "Connection" && k != "Transfer-Encoding" && k != "X-Cache" {
						if len(v) > 0 {
							response.Headers[k] = v[0]
						}
					}
				}

				// Сохраняем в Redis асинхронно (не блокируем ответ)
				// Используем контекст с таймаутом, чтобы не зависеть от завершения запроса
				go func() {
					cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					data, err := response.Marshal()
					if err == nil {
						if err := redisClient.Set(cacheCtx, cacheKey, data, ttl).Err(); err != nil {
							log.Warn("Failed to cache response", map[string]interface{}{
								"error": err.Error(),
								"key":   cacheKey,
							})
						}
					}
				}()
			}

			// X-Cache заголовок уже установлен в recorder
		})
	}
}

// generateCacheKey генерирует ключ кэша на основе URL и query параметров
func generateCacheKey(r *http.Request) string {
	key := r.URL.Path + "?" + r.URL.RawQuery
	hash := sha256.Sum256([]byte(key))
	return "cache:" + hex.EncodeToString(hash[:])
}

// responseRecorder записывает ответ для кэширования
type responseRecorder struct {
	http.ResponseWriter
	statusCode  int
	body        *bytes.Buffer
	headers     http.Header
	wroteHeader bool
}

func (r *responseRecorder) Header() http.Header {
	return r.headers
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.statusCode = code
	r.wroteHeader = true

	// Копируем заголовки в оригинальный ResponseWriter
	for k, v := range r.headers {
		r.ResponseWriter.Header()[k] = v
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// cachedResponse структура для хранения кэшированного ответа
type cachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

// Marshal сериализует cachedResponse в JSON
func (c *cachedResponse) Marshal() ([]byte, error) {
	// Простая сериализация через JSON (можно улучшить)
	return json.Marshal(c)
}

// Unmarshal десериализует cachedResponse из JSON
func (c *cachedResponse) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}
