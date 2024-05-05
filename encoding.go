package ggd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

type EncodingFormatter func(HexEncoding) string
type encoderOption func(d *hexEncoder) error

type hexEncoder struct {
	chunkSize int
	output    io.Writer
	input     io.Reader
	formatter EncodingFormatter
}

func DefaultEncFormatter(hx HexEncoding) string {
	hexCodes := ""
	for _, hb := range hx.HexCodes {
		hexCodes += hb.String()
	}
	return fmt.Sprintf("%s", hexCodes)
}

func NewEncoder(opt ...encoderOption) (*hexEncoder, error) {
	e := &hexEncoder{
		chunkSize: DefaultBufSize,
		output:    os.Stdout,
		input:     os.Stdin,
		formatter: DefaultEncFormatter,
	}

	for _, o := range opt {
		err := o(e)
		if err != nil {
			return &hexEncoder{}, err
		}
	}

	return e, nil
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

func (he hexEncoder) Encode() error {
	scanner := bufio.NewScanner(he.input)

	scanner.Split(encoderSplitFunc(he))
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
		fmt.Fprintln(he.output, he.formatter(hex))
	}

	return nil
}
