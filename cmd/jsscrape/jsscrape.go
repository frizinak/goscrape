package main

import (
	"flag"
	"os"
	"strings"

	"github.com/frizinak/goscrape/cmd"
	"github.com/gopherjs/gopherjs/js"
)

type Console struct {
	*js.Object
	method string
}

func NewConsole(method string) *Console {
	c := &Console{Object: js.Global.Get("console")}
	c.method = method
	return c
}

func (c *Console) Write(p []byte) (int, error) {
	c.WriteString(string(p))
	return len(p), nil
}

func (c *Console) WriteString(p string) {
	c.Call(c.method, p)
}

func run(name string, args []string) error {
	stderr := NewConsole("error")
	flags := flag.NewFlagSet(
		name,
		flag.ExitOnError,
	)

	err := cmd.Cmd(flags, args, stderr)
	if err != nil {
		stderr.WriteString(err.Error())
	}

	return err
}

func main() {
	js.Global.Set("run", func(args []string) {
		go run("run", args)
	})

	if !js.Global.Get("window").Bool() {
		jsargs := js.Global.Get("process").Get("argv")
		args := make([]string, 0, jsargs.Length())
		for i := 0; i < jsargs.Length(); i++ {
			args = append(args, jsargs.Index(i).String())
		}

		if err := run(strings.Join(args[0:2], " "), args[2:]); err != nil {
			os.Exit(1)
		}
	}
}
