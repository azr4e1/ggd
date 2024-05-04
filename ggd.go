package ggd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const (
	DefaultBufSize = 16
	DefaultGroups  = 2
)

var validHexDigits = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}

type Formatter func(HexDump) string
type option func(d *hexDumper) error

type hexDumper struct {
	chunkSize int
	output    io.Writer
	input     io.Reader
	formatter Formatter
}

type HexDump struct {
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

func SingleByteDump(b byte) HexByte {
	firstHex, _ := ConvertToHexadecimal((b & 0b11110000) >> 4)
	secondHex, _ := ConvertToHexadecimal(b & 0b00001111)

	return HexByte{
		first:  firstHex,
		second: secondHex,
	}
}

func DefaultFormatter(hx HexDump) string {
	hexCodes := ""
	for _, hb := range hx.HexCodes {
		hexCodes += hb.String()
	}
	return fmt.Sprintf("%s", hexCodes)
}

func NewDumper(opt ...option) (*hexDumper, error) {
	d := &hexDumper{
		chunkSize: DefaultBufSize,
		output:    os.Stdout,
		input:     os.Stdin,
		formatter: DefaultFormatter,
	}

	for _, o := range opt {
		err := o(d)
		if err != nil {
			return &hexDumper{}, err
		}
	}

	return d, nil
}

func DumperChunkSize(bs int) option {
	return func(d *hexDumper) error {
		if bs < 0 {
			return errors.New("no columns")
		}
		d.chunkSize = bs

		return nil
	}
}

func DumperInput(input io.Reader) option {
	return func(d *hexDumper) error {
		if input == nil {
			return errors.New("nil input")
		}
		d.input = input

		return nil
	}
}

func DumperOutput(output io.Writer) option {
	return func(d *hexDumper) error {
		if output == nil {
			return errors.New("nil output")
		}
		d.output = output

		return nil
	}
}

func DumperFormatter(f Formatter) option {
	return func(d *hexDumper) error {

		d.formatter = f

		return nil
	}
}

func newSplitFunc(hd hexDumper) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		chunkSize := hd.chunkSize

		for i := 0; i < len(data); i++ {
			if i == chunkSize-1 {
				return chunkSize, data[:chunkSize], nil
			}
		}

		if atEOF && len(data) > 0 {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
}

// TODO: read as a stream of bytes, not slice of bytes
func (hd hexDumper) Dump() error {
	scanner := bufio.NewScanner(hd.input)

	scanner.Split(newSplitFunc(hd))
	offset := 0

	for scanner.Scan() {
		input := scanner.Bytes()
		hex := HexDump{
			Input:    input,
			Offset:   offset,
			HexCodes: []HexByte{},
		}
		for _, b := range input {
			hex.HexCodes = append(hex.HexCodes, SingleByteDump(b))
		}

		offset += len(input)
		fmt.Fprintln(hd.output, hd.formatter(hex))
	}

	return nil
}
