package vite

import (
	"encoding/hex"
	"fmt"

	"github.com/vitelabs/go-vite/crypto/ed25519"
)

func SignData(privateKey string, message string) error {

	privKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return err
	}

	data, err := hex.DecodeString(message)
	if err != nil {
		return err
	}

	signature := ed25519.Sign(privKey, data)
	fmt.Println(hex.EncodeToString(signature))
	
	return nil
}
