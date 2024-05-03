package cmdline

import (
	// "github.com/azr4e1/ggd"

	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/azr4e1/ggd"
)

const (
	ErrorFlag = 1
)

const (
	PrintableMinASCII = 32
	PrintableMaxASCII = unicode.MaxASCII
)

type option func(*cmdDumper) error
type Formatter func(ggd.HexDump) string

type cmdDumper struct {
	Input     io.Reader
	Output    io.Writer
	Columns   int
	Groups    int
	Color     bool
	Formatter Formatter
	files     []io.Reader
}

func IsAscii(b byte) bool {
	return b < PrintableMaxASCII && b >= PrintableMinASCII
}

func DefaultFormat(hx ggd.HexDump) string {
	normalizedInput := []byte{}
	for _, b := range hx.Input {
		if !IsAscii(b) || b == '\n' || b == '\t' {
			normalizedInput = append(normalizedInput, '.')
			continue
		}
		normalizedInput = append(normalizedInput, b)
	}

	return fmt.Sprintf("%d:\t%s\t%s", hx.Offset, strings.Join(hx.Output, " "), normalizedInput)
}

func NewCmdDumper(opts ...option) (*cmdDumper, error) {
	cmdD := &cmdDumper{
		Input:     os.Stdin,
		Output:    os.Stdout,
		Columns:   16,
		Groups:    2,
		Color:     true,
		Formatter: DefaultFormat,
	}

	for _, o := range opts {
		err := o(cmdD)
		if err != nil {
			return &cmdDumper{}, err
		}
	}

	return cmdD, nil
}

func (cd cmdDumper) Format(hx []ggd.HexDump) []string {
	formatted := []string{}
	for _, h := range hx {
		formatted = append(formatted, cd.Formatter(h))
	}

	return formatted
}

func WithInput(r io.Reader) option {
	return func(cd *cmdDumper) error {
		cd.Input = r
		return nil
	}
}

func WithOutput(w io.Writer) option {
	return func(cd *cmdDumper) error {
		cd.Output = w
		return nil
	}
}

func WithColumns(c int) option {
	return func(cd *cmdDumper) error {
		if c <= 0 {
			return errors.New("invalid number of columns")
		}
		cd.Columns = c
		return nil
	}
}

func WithGroups(g int) option {
	return func(cd *cmdDumper) error {
		if g <= 0 {
			return errors.New("invalid number of groups")
		}
		cd.Groups = g
		return nil
	}
}

func WithFormat(f Formatter) option {
	return func(cd *cmdDumper) error {
		cd.Formatter = f
		return nil
	}
}

func WithColor(c bool) option {
	return func(cd *cmdDumper) error {
		cd.Color = c
		return nil
	}
}

func WithInputFromArgs(args []string) option {
	return func(cd *cmdDumper) error {
		if len(args) < 1 {
			return nil
		}
		cd.files = make([]io.Reader, len(args))
		for i, path := range args {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			cd.files[i] = f
		}
		cd.Input = io.MultiReader(cd.files...)
		return nil
	}
}

// func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error)

func (cd *cmdDumper) Dump() error {
	// input := bufio.NewScanner(cd.Input)
	// // type SplitFunc func(data []byte, atEOF bool) (advance int, token []byte, err error)
	// input.Split(SplitFunc)
	//
	// for input.Scan() {
	// 	data := input.Bytes()
	// }
	for _, f := range cd.files {
		defer f.(io.Closer).Close()
	}

	data, err := io.ReadAll(cd.Input)
	if err != nil {
		return err
	}
	dumper, err := ggd.NewDumper(ggd.DumperGroups(cd.Groups), ggd.DumperColumns(cd.Columns))
	if err != nil {
		return err
	}
	dump := dumper.Dump(data)

	fmt.Fprint(cd.Output, strings.Join(cd.Format(dump), "\n"))

	return nil
}

func Main() int {
	groups := flag.Int("groups", 2, "number of hex codes in a single group")
	columns := flag.Int("columns", 16, "number of hex codes in a single line")
	color := flag.Bool("color", true, "colored output")
	outputName := flag.String("output", "", "output file")
	flag.Parse()

	var output io.Writer = os.Stdout
	if *outputName != "" {
		outputFile, err := os.Create(*outputName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return ErrorFlag
		}
		defer outputFile.Close()

		output = outputFile
	}

	dumper, err := NewCmdDumper(WithColor(*color), WithColumns(*columns), WithGroups(*groups), WithInputFromArgs(flag.Args()), WithOutput(output))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ErrorFlag
	}

	dumper.Dump()

	return 0
}
