package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/daviddengcn/go-colortext"
	anko_core "github.com/mattn/anko/builtins/core"
	anko_http "github.com/mattn/anko/builtins/http"
	anko_io "github.com/mattn/anko/builtins/io"
	anko_json "github.com/mattn/anko/builtins/json"
	anko_os "github.com/mattn/anko/builtins/os"
	anko_url "github.com/mattn/anko/builtins/url"
	"github.com/mattn/anko/parser"
	"github.com/mattn/anko/vm"
	"github.com/mattn/go-isatty"
	"io/ioutil"
	"os"
	"reflect"
)

const version = "0.0.1"

var e = flag.String("e", "", "One line of program")
var v = flag.Bool("v", false, "Display version")

var istty = isatty.IsTerminal(os.Stdout.Fd())

func colortext(color ct.Color, bright bool, f func()) {
	if istty {
		ct.ChangeColor(color, bright, ct.None, false)
		f()
		ct.ResetColor()
	} else {
		f()
	}
}

func main() {
	flag.Parse()
	if *v {
		fmt.Println(version)
		os.Exit(0)
	}

	env := vm.NewEnv()

	anko_core.Import(env)
	anko_http.Import(env)
	anko_url.Import(env)
	anko_json.Import(env)
	anko_os.Import(env)
	anko_io.Import(env)

	var code string
	var b []byte
	var reader *bufio.Reader
	var following bool

	repl := flag.NArg() == 0 && *e == ""

	env.Define("args", reflect.ValueOf(flag.Args()))

	if repl {
		reader = bufio.NewReader(os.Stdin)
	} else {
		if *e != "" {
			b = []byte(*e)
		} else {
			var err error
			b, err = ioutil.ReadFile(flag.Arg(0))
			if err != nil {
				colortext(ct.Red, false, func() {
					fmt.Fprintln(os.Stderr, err)
				})
				os.Exit(1)
			}
			env.Define("args", reflect.ValueOf(flag.Args()[1:]))
		}
	}

	for {
		if repl {
			colortext(ct.Green, true, func() {
				if following {
					fmt.Print("  ")
				} else {
					fmt.Print("> ")
				}
			})
			b, _, err := reader.ReadLine()
			if err != nil {
				break
			}
			if len(b) == 0 {
				continue
			}
			code += string(b)
		} else {
			code = string(b)
		}

		scanner := new(parser.Scanner)
		scanner.Init(code)
		stmts, err := parser.Parse(scanner)

		v := vm.NilValue

		if repl {
			if following {
				continue
			}
			if e, ok := err.(*parser.Error); ok && e.Pos().Column == len(b) {
				following = true
				continue
			}
			if err == nil {
				following = false
				code = ""
			}
		}
		if err == nil {
			v, err = vm.RunStmts(stmts, env)
		}

		if err != nil {
			colortext(ct.Red, false, func() {
				if e, ok := err.(*vm.Error); ok {
					fmt.Fprintf(os.Stderr, "typein:%d: %s\n", e.Pos().Line, err)
				} else if e, ok := err.(*parser.Error); ok {
					fmt.Fprintf(os.Stderr, "typein:%d: %s\n", e.Pos().Line, err)
				} else {
					fmt.Fprintln(os.Stderr, err)
				}
			})

			if repl {
				continue
			} else {
				os.Exit(1)
			}
		} else {
			if repl {
				colortext(ct.Black, true, func() {
					if v == vm.NilValue {
						fmt.Println("nil")
					} else {
						fmt.Println(v.Interface())
					}
				})
			} else {
				break
			}
		}
	}
}
