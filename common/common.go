package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

var IsDebug bool

func Print(v ...interface{}) {
	if IsDebug {
		fmt.Println(v)
	}
}
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func BytesToHex(bytes []byte) string {
	hex := hex.EncodeToString(bytes)
	return hex
}

func HexToBytes(hexString string) []byte {
	bytes, _ := hex.DecodeString(hexString)
	return bytes
}
