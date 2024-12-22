package models

type GetLRU struct {
	Keys   []string      `json:"keys"`
	Values []interface{} `json:"values"`
}

type LRUResponse struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt int64       `json:"expires"`
}

type PutRequest struct {
	Key        string      `json:"key" validate:"required"`
	Value      interface{} `json:"value" validate:"required"`
	TTLSeconds int         `json:"ttl_seconds,omitempty" validate:"number,gte=0"`
}
