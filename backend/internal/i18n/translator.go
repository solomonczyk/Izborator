package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Translator переводчик для мультиязычности
type Translator struct {
	locales map[string]map[string]string
}

// NewTranslator создаёт новый переводчик и загружает локали
func NewTranslator(localesDir string) (*Translator, error) {
	t := &Translator{
		locales: make(map[string]map[string]string),
	}

	// Поддерживаемые языки
	languages := []string{"en", "sr", "ru", "hu", "zh"}

	for _, lang := range languages {
		filePath := filepath.Join(localesDir, fmt.Sprintf("%s.json", lang))
		data, err := os.ReadFile(filePath)
		if err != nil {
			// Если файл не найден, создаём пустую локаль
			t.locales[lang] = make(map[string]string)
			continue
		}

		var locale map[string]string
		if err := json.Unmarshal(data, &locale); err != nil {
			return nil, fmt.Errorf("failed to parse locale %s: %w", lang, err)
		}

		t.locales[lang] = locale
	}

	// Если английский не загружен, создаём пустой
	if _, ok := t.locales["en"]; !ok {
		t.locales["en"] = make(map[string]string)
	}

	return t, nil
}

// T переводит ключ на указанный язык, с fallback на английский
func (t *Translator) T(lang, key string) string {
	// Нормализуем язык (например, "sr-RS" -> "sr")
	lang = normalizeLang(lang)

	// Пытаемся найти перевод
	if locale, ok := t.locales[lang]; ok {
		if val, ok := locale[key]; ok && val != "" {
			return val
		}
	}

	// Fallback на английский
	if locale, ok := t.locales["en"]; ok {
		if val, ok := locale[key]; ok && val != "" {
			return val
		}
	}

	// Если даже английского нет, возвращаем ключ
	return key
}

// normalizeLang нормализует код языка (например, "sr-RS" -> "sr", "en-US" -> "en")
func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))

	// Убираем регион (например, "sr-RS" -> "sr")
	if idx := strings.Index(lang, "-"); idx > 0 {
		lang = lang[:idx]
	}

	// Проверяем, поддерживается ли язык
	supported := map[string]bool{
		"en": true,
		"sr": true,
		"ru": true,
		"hu": true,
		"zh": true,
	}

	if supported[lang] {
		return lang
	}

	// Fallback на английский
	return "en"
}

// GetSupportedLanguages возвращает список поддерживаемых языков
func (t *Translator) GetSupportedLanguages() []string {
	return []string{"en", "sr", "ru", "hu", "zh"}
}
