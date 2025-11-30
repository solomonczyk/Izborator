package middleware

import (
	"context"
	"net/http"
	"strings"
)

// LangContextKey ключ для хранения языка в контексте
type LangContextKey string

const LangKey LangContextKey = "lang"

// DetectLanguage middleware для определения языка из запроса
func DetectLanguage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := detectLanguageFromRequest(r)
		
		// Сохраняем язык в контексте
		ctx := context.WithValue(r.Context(), LangKey, lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// detectLanguageFromRequest определяет язык из запроса
// Приоритет: query param > Accept-Language header > default "en"
func detectLanguageFromRequest(r *http.Request) string {
	// 1. Проверяем query параметр ?lang=xx
	if lang := r.URL.Query().Get("lang"); lang != "" {
		return normalizeLang(lang)
	}

	// 2. Проверяем Accept-Language header
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		// Парсим Accept-Language (например, "sr-RS,en;q=0.9")
		langs := parseAcceptLanguage(acceptLang)
		for _, lang := range langs {
			normalized := normalizeLang(lang)
			if isSupported(normalized) {
				return normalized
			}
		}
	}

	// 3. Fallback на английский
	return "en"
}

// parseAcceptLanguage парсит Accept-Language header и возвращает список языков по приоритету
func parseAcceptLanguage(header string) []string {
	// Упрощённый парсинг: берём первый язык из списка
	parts := strings.Split(header, ",")
	if len(parts) > 0 {
		lang := strings.TrimSpace(parts[0])
		// Убираем quality (например, "sr-RS;q=0.9" -> "sr-RS")
		if idx := strings.Index(lang, ";"); idx > 0 {
			lang = lang[:idx]
		}
		return []string{lang}
	}
	return []string{}
}

// normalizeLang нормализует код языка
func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	
	// Убираем регион (например, "sr-RS" -> "sr")
	if idx := strings.Index(lang, "-"); idx > 0 {
		lang = lang[:idx]
	}

	return lang
}

// isSupported проверяет, поддерживается ли язык
func isSupported(lang string) bool {
	supported := map[string]bool{
		"en": true,
		"sr": true,
		"ru": true,
		"hu": true,
		"zh": true,
	}
	return supported[lang]
}

// GetLangFromContext получает язык из контекста
func GetLangFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(LangKey).(string); ok {
		return lang
	}
	return "en"
}

