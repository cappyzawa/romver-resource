package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	resource "github.com/cappyzawa/romver-resource"
	"github.com/cappyzawa/romver-resource/driver"
)

// Check represents check command stream
type Check struct {
	InStream  io.Reader
	ErrStream io.Writer
	OutStream io.Writer
}

func (c *Check) Execute(args []string) int {
	var req resource.CheckRequest
	if err := json.NewDecoder(c.InStream).Decode(&req); err != nil {
		return c.fatal("decoding request", err)
	}

	driver, err := driver.FromSource(req.Source)
	if err != nil {
		return c.fatal("constructing driver", err)
	}
	var cursor string
	if req.Version != nil {
		cursor = req.Version.Number
	}

	versions, err := driver.Check(cursor)
	if err != nil {
		return c.fatal("checking for new versions", err)
	}

	var res resource.CheckResponse
	for _, v := range versions {
		res = append(res, resource.Version{
			Number: v,
		})
	}
	if err := json.NewEncoder(c.OutStream).Encode(res); err != nil {
		return c.fatal("encoding response", err)
	}
	return 0
}

func (c *Check) fatal(doing string, err error) int {
	fmt.Fprintf(c.ErrStream, "error %s: %v", doing, err)
	return 1
}

func main() {
	c := &Check{
		InStream:  os.Stdin,
		ErrStream: os.Stderr,
		OutStream: os.Stdout,
	}
	os.Exit(c.Execute(os.Args))
}
