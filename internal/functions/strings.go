package functions

import (
	"encoding/base64"
	"encoding/json"
)

func base64encode(v any) (any, error) {
	s := v.(string)
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

func base64decode(v any) (any, error) {
	s := v.(string)
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err.Error(), err
	}
	return string(data), nil
}

// toJson encodes an item into a pretty (indented) JSON string
func toJson(v any) (any, error) {
	output, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
