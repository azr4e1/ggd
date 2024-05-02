package ggd

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var validHexDigits = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}

type hexByte struct {
	First  byte
	Second byte
}

func (hx hexByte) String() string {
	return strings.ToLower(fmt.Sprintf("%s%s", string(hx.First), string(hx.Second)))
}

func NewHex(first, second byte) (hexByte, error) {
	if !slices.Contains(validHexDigits, first) {
		return hexByte{}, fmt.Errorf("%v is not a valid hex digit", first)
	}
	if !slices.Contains(validHexDigits, second) {
		return hexByte{}, fmt.Errorf("%v is not a valid hex digit", second)
	}

	return hexByte{first, second}, nil
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

func SingleByteDump(b byte) hexByte {
	firstHex, _ := ConvertToHexadecimal((b & 0b11110000) >> 4)
	secondHex, _ := ConvertToHexadecimal(b & 0b00001111)

	return hexByte{
		First:  firstHex,
		Second: secondHex,
	}
}

func HexDump(bs []byte) []hexByte {
	dump := make([]hexByte, len(bs))
	for _, el := range bs {
		dump = append(dump, SingleByteDump(el))
	}

	return dump
}
