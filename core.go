package ggd

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	DefaultBufSize = 16
	DefaultGroups  = 2
)

var validHexDigits = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

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

func NewHex(first, second byte) (HexByte, error) {
	if !slices.Contains(validHexDigits, first) {
		return HexByte{}, fmt.Errorf("%v is not a valid hex digit", first)
	}
	if !slices.Contains(validHexDigits, second) {
		return HexByte{}, fmt.Errorf("%v is not a valid hex digit", second)
	}

	return HexByte{first, second}, nil
}

func ConvertToHexadecimal(b byte) (byte, error) {
	switch {
	case b <= 9:
		return '0' + b, nil
	case b <= 15:
		return 'a' + (b - 10), nil
	}

	return 0, errors.New("Not a valid hexadecimal value")
}

func SingleByteEncode(b byte) HexByte {
	firstHex, _ := ConvertToHexadecimal((b & 0b11110000) >> 4)
	secondHex, _ := ConvertToHexadecimal(b & 0b00001111)

	return HexByte{
		first:  firstHex,
		second: secondHex,
	}
}

func ConvertToByte(b byte) (byte, error) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', nil
	case b >= 'a' && b <= 'f':
		return 10 + b - 'a', nil
	}

	return 0, errors.New("Not a valid hexadecimal value")
}

func (hx HexByte) Byte() byte {
	first := hx.first
	second := hx.second

	firstB, _ := ConvertToByte(first)
	secondB, _ := ConvertToByte(second)

	return (firstB << 4) + secondB
}
