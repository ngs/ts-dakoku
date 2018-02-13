package app

import (
	"crypto/rand"
	"encoding/binary"
	"strconv"
)

func RandomString(len int) string {
	var n uint64
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	return strconv.FormatUint(n, len)
}
