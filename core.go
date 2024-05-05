package ggd

import (
	"fmt"
	"strings"
)

const (
	DefaultBufSize = 16
	DefaultGroups  = 2
)

var validHexDigits = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}

type HexEncoding struct {
	Input    []byte
	Offset   int
	HexCodes []HexByte
}

type HexByte struct {
	first  byte
	second byte
}

func (hx HexByte) String() string {
	return strings.ToLower(fmt.Sprintf("%s%s", string(hx.first), string(hx.second)))
}
