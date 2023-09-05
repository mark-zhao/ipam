package aeser

import (
	"encoding/hex"
	"fmt"
)

func example() {
	hexKey := "6c1acf9ad6f12ff7e3c5b94df9f9ef329996b6ea7d148afafe76765d42d0a876"
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(err)
	}
	encryptResult, err := AESEncrypt([]byte("polaris@studygolang"), key)
	if err != nil {
		panic(err)
	}
	result := hex.EncodeToString(encryptResult)
	fmt.Println(result)
	encryptResult, err = hex.DecodeString(result)
	if err != nil {
		panic(err)
	}
	origData, err := AESDecrypt(encryptResult, key)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(origData))
}
