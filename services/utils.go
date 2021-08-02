package services

import (
	"encoding/base64"
	"encoding/json"

	"github.com/vitelabs/go-vite/rpcapi/api"
)

// *JSONMap functions are needed because `types.MarshalMap/types.UnmarshalMap`
// does not respect custom JSON marshalers.

// marshalJSONMap converts an interface into a map[string]interface{}.
func marshalJSONMap(i interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return m, nil
}

// unmarshalJSONMap converts map[string]interface{} into a interface{}.
func unmarshalJSONMap(m map[string]interface{}, i interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, i)
}

// Decodes an account block from a base64 encoded string
func decodeAccountBlockFromBase64(data string) (*api.AccountBlock, error) {
	var accountBlock api.AccountBlock
	jsonData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonData, &accountBlock); err != nil {
		return nil, err
	}
	return &accountBlock, nil
}

// Encodes an account block to a base64 string
func encodeAccountBlockToBase64(accountBlock api.AccountBlock) (string, error) {
	jsonData, err := json.Marshal(accountBlock)
	if err != nil {
		return "", err
	}
	data := base64.StdEncoding.EncodeToString(jsonData)
	return data, nil
}
