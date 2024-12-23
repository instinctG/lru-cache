// Package models содержит модели данных, используемые в приложении.
package models

// GetLRU представляет структуру для возврата всех ключей и значений из LRU-кэша.
type GetLRU struct {
	Keys   []string      `json:"keys"`   // Список ключей, хранящихся в кэше.
	Values []interface{} `json:"values"` // Список значений, соответствующих ключам.
}

// LRUResponse представляет структуру ответа для операций с отдельным элементом LRU-кэша.
type LRUResponse struct {
	Key       string      `json:"key"`     // Ключ элемента.
	Value     interface{} `json:"value"`   // Значение элемента.
	ExpiresAt int64       `json:"expires"` // Время истечения срока действия элемента в формате Unix Time.
}

// PutRequest представляет структуру запроса для добавления или обновления элемента в LRU-кэше.
type PutRequest struct {
	Key        string      `json:"key" validate:"required"`                       // Ключ элемента (обязательное поле).
	Value      interface{} `json:"value" validate:"required"`                     // Значение элемента (обязательное поле).
	TTLSeconds int         `json:"ttl_seconds,omitempty" validate:"number,gte=0"` // Время жизни элемента в секундах (необязательное поле, должно быть >= 0).
}
