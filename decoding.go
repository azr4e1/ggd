package ggd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

type DecodingFormatter func(string) ([]HexByte, error)
type decoderOption func(*hexDecoder) error

type hexDecoder struct {
	output    io.Writer
	input     io.Reader
	formatter DecodingFormatter
}

func DefaultDecFormatter(s string) ([]HexByte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("\"%s\" doesn't represent a correctly encoded byte sequence (odd number of hex values)\n", s)
	}

	hexBytes := []HexByte{}
	for i := 0; i < len(s)-1; i += 2 {
		first := s[i]
		second := s[i+1]
		hex, err := NewHex(first, second)
		if err != nil {
			return nil, err
		}

		hexBytes = append(hexBytes, hex)
	}

	return hexBytes, nil
}

func DecoderInput(input io.Reader) decoderOption {
	return func(d *hexDecoder) error {
		if input == nil {
			return errors.New("nil input")
		}
		d.input = input

		return nil
	}
}

func DecoderOutput(output io.Writer) decoderOption {
	return func(d *hexDecoder) error {
		if output == nil {
			return errors.New("nil output")
		}
		d.output = output

		return nil
	}
}

func DecoderFormatter(f DecodingFormatter) decoderOption {
	return func(d *hexDecoder) error {

		d.formatter = f

		return nil
	}
}

func NewDecoder(opt ...decoderOption) (*hexDecoder, error) {
	d := &hexDecoder{
		input:     os.Stdin,
		output:    os.Stdout,
		formatter: DefaultDecFormatter,
	}

	for _, o := range opt {
		err := o(d)
		if err != nil {
			return &hexDecoder{}, err
		}
	}

	return d, nil
}

func (hd hexDecoder) Decode() error {
	scanner := bufio.NewScanner(hd.input)

	for scanner.Scan() {
		input := scanner.Text()
		hexBytes, err := hd.formatter(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		decodedHex := []byte{}
		for _, hex := range hexBytes {
			decodedHex = append(decodedHex, hex.Byte())
		}
		fmt.Fprintf(hd.output, "%s", string(decodedHex))
	}

	return nil
}
