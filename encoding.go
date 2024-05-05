package ggd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
)

type EncodingFormatter func(HexEncoding) string
type encoderOption func(d *hexEncoder) error

type hexEncoder struct {
	chunkSize int
	output    io.Writer
	input     io.Reader
	formatter EncodingFormatter
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

func DefaultFormatter(hx HexEncoding) string {
	hexCodes := ""
	for _, hb := range hx.HexCodes {
		hexCodes += hb.String()
	}
	return fmt.Sprintf("%s", hexCodes)
}

func NewEncoder(opt ...encoderOption) (*hexEncoder, error) {
	d := &hexEncoder{
		chunkSize: DefaultBufSize,
		output:    os.Stdout,
		input:     os.Stdin,
		formatter: DefaultFormatter,
	}

	for _, o := range opt {
		err := o(d)
		if err != nil {
			return &hexEncoder{}, err
		}
	}

	return d, nil
}

func EncoderChunkSize(bs int) encoderOption {
	return func(d *hexEncoder) error {
		if bs < 0 {
			return errors.New("no columns")
		}
		d.chunkSize = bs

		return nil
	}
}

func EncoderInput(input io.Reader) encoderOption {
	return func(d *hexEncoder) error {
		if input == nil {
			return errors.New("nil input")
		}
		d.input = input

		return nil
	}
}

func EncoderOutput(output io.Writer) encoderOption {
	return func(d *hexEncoder) error {
		if output == nil {
			return errors.New("nil output")
		}
		d.output = output

		return nil
	}
}

func EncoderFormatter(f EncodingFormatter) encoderOption {
	return func(d *hexEncoder) error {

		d.formatter = f

		return nil
	}
}

func encoderSplitFunc(hd hexEncoder) bufio.SplitFunc {
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
func (hd hexEncoder) Encode() error {
	scanner := bufio.NewScanner(hd.input)

	scanner.Split(encoderSplitFunc(hd))
	offset := 0

	for scanner.Scan() {
		input := scanner.Bytes()
		hex := HexEncoding{
			Input:    input,
			Offset:   offset,
			HexCodes: []HexByte{},
		}
		for _, b := range input {
			hex.HexCodes = append(hex.HexCodes, SingleByteEncode(b))
		}

		offset += len(input)
		fmt.Fprintln(hd.output, hd.formatter(hex))
	}

	return nil
}
