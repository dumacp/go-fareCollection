package qr

import (
	"encoding/json"
	"fmt"

	"github.com/dumacp/go-fareCollection/pkg/messages"
)

const (
	EQPM = "EQPM"
	AQPM = "AQPM"
)

func paym(value json.RawMessage) (map[string]string, map[string]*messages.Value, error) {
	var mapp map[string]interface{}
	if err := json.Unmarshal(value, &mapp); err != nil {
		return nil, nil, fmt.Errorf("QR error: %w", err)
	}
	raw := make(map[string]string)
	data := make(map[string]*messages.Value)
	for k, v := range mapp {
		switch value := v.(type) {
		case float64:
			raw[k] = fmt.Sprintf("%d", int32(value))
			data[k] = &messages.Value{Data: &messages.Value_IntValue{IntValue: int32(value)}}
		case int:
			raw[k] = fmt.Sprintf("%d", value)
			data[k] = &messages.Value{Data: &messages.Value_IntValue{IntValue: int32(value)}}
		case uint:
			raw[k] = fmt.Sprintf("%d", value)
			data[k] = &messages.Value{Data: &messages.Value_UintValue{UintValue: uint32(value)}}
		case int64:
			raw[k] = fmt.Sprintf("%d", value)
			data[k] = &messages.Value{Data: &messages.Value_Int64Value{Int64Value: int64(value)}}
		case uint64:
			raw[k] = fmt.Sprintf("%d", value)
			data[k] = &messages.Value{Data: &messages.Value_Uint64Value{Uint64Value: uint64(value)}}
		case string:
			raw[k] = value
			data[k] = &messages.Value{Data: &messages.Value_StringValue{StringValue: value}}
		case []byte:
			raw[k] = fmt.Sprintf("%s", value)
			data[k] = &messages.Value{Data: &messages.Value_BytesValue{BytesValue: value}}
		}
	}
	return raw, data, nil
}
