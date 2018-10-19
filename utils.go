package address

import (
	"bytes"
	"encoding/binary"
)

func IntToHex(data int64) []byte {
	result := new(bytes.Buffer)
	binary.Write(result, binary.BigEndian, data)
	return result.Bytes()
}
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
