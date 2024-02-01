package provider

import (
	"encoding/json"
	"fmt"
)

type RateLimitError struct {
	message string
	kv      map[string]any
}

func NewRateLimitError(message string, kvs ...any) RateLimitError {
	err := RateLimitError{
		message: message,
	}

	if len(kvs) > 0 {
		err.kv = parseKVs(kvs)
	}

	return err
}

func (r RateLimitError) Error() string {
	message := r.message

	for k, v := range r.kv {
		message += fmt.Sprintf(" %s=%v", k, v)
	}

	return message
}

func (r RateLimitError) KVs() []any {
	kvs := make([]any, 0, len(r.kv)*2)
	for k, v := range r.kv {
		kvs = append(kvs, k, v)
	}

	return kvs
}

func parseKVs(kvs []any) map[string]any {
	kv := make(map[string]any, len(kvs))
	for i := 0; i < len(kvs); i += 2 {
		key, ok := kvs[i].(string)
		if !ok {
			keyBytes, err := json.Marshal(kvs[i])
			if err != nil {
				key = "<invalid>"
			} else {
				key = fmt.Sprintf("<non-string: %s>", keyBytes)
			}
		}

		var value any
		if i+1 >= len(kvs) {
			value = "<not provided>"
		} else {
			value = kvs[i+1]
		}

		kv[key] = value
	}

	return kv
}
