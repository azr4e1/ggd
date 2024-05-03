package ggd

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	DefaultColumns = 16
	DefaultGroups  = 2
)

var validHexDigits = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}

type hexDumper struct {
	columns int
	groups  int
}

type HexDump struct {
	Input  []byte
	Output []string
	Offset int
	Raw    []hexByte
}

type option func(d *hexDumper) error

type hexByte struct {
	first  byte
	second byte
}

func (hx hexByte) String() string {
	return strings.ToLower(fmt.Sprintf("%s%s", string(hx.first), string(hx.second)))
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
		first:  firstHex,
		second: secondHex,
	}
}

func NewDumper(opt ...option) (*hexDumper, error) {
	d := &hexDumper{
		columns: DefaultColumns,
		groups:  DefaultGroups,
	}

	for _, o := range opt {
		err := o(d)
		if err != nil {
			return &hexDumper{}, err
		}
	}

	return d, nil
}

func DumperColumns(c int) option {
	return func(d *hexDumper) error {
		if c < 0 {
			return errors.New("no columns")
		}
		d.columns = c

		return nil
	}
}

func DumperGroups(g int) option {
	return func(d *hexDumper) error {
		if g <= 0 {
			return errors.New("no groups")
		}
		d.groups = g

		return nil
	}
}

// TODO: read as a stream of bytes, not slice of bytes
func (hd hexDumper) Dump(bs []byte) []HexDump {
	index := 0
	length := len(bs)
	dumps := []HexDump{}
	groups := []string{}
	currDump := HexDump{Offset: 0}
	currBytes := []byte{}
	currHexes := []hexByte{}
	for index < length {
		curr := ""
		for g := 0; g < hd.groups; g++ {
			b := bs[index]
			currBytes = append(currBytes, b)

			hex := SingleByteDump(b)
			currHexes = append(currHexes, hex)

			curr += hex.String()

			index++
			if index%hd.columns == 0 || index >= length {
				break
			}
		}
		groups = append(groups, curr)

		if index%hd.columns == 0 || index >= length {
			currDump.Input = currBytes
			currDump.Raw = currHexes
			currDump.Output = groups

			dumps = append(dumps, currDump)

			currDump = HexDump{Offset: index}
			currBytes = []byte{}
			currHexes = []hexByte{}
			groups = []string{}
		}

	}
	return dumps
}
