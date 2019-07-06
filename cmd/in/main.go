package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	resource "github.com/cappyzawa/romver-resource"
)

// In represents in command stream
type In struct {
	InStream  io.Reader
	ErrStream io.Writer
	OutStream io.Writer
}

// Execute executes in command
func (i *In) Execute(args []string) int {
	if len(args) < 2 {
		fmt.Fprintf(i.ErrStream, "usage: %s <destination>", args[0])
		return 1
	}

	destDir := args[1]

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return i.fatal("creating destination", err)
	}

	var req resource.InRequest
	if err := json.NewDecoder(i.InStream).Decode(&req); err != nil {
		return i.fatal("decoding request", err)
	}

	version := req.Version.Number
	if req.Params.Bump {
		versionInt, err := strconv.Atoi(version)
		if err != nil {
			return i.fatal("coverting version to int", err)
		}
		version = strconv.Itoa(versionInt + 1)
	}

	versionFileNames := []string{"number", "version"}
	for _, fileName := range versionFileNames {
		numberFile, err := os.Create(filepath.Join(destDir, fileName))
		if err != nil {
			return i.fatal("opening number file", err)
		}
		defer numberFile.Close()

		if _, err := fmt.Fprintf(numberFile, "%s", version); err != nil {
			return i.fatal("writing number file", err)
		}
	}

	res := resource.InResponse{
		Version: req.Version,
		Metadata: []resource.MetadataField{
			{Name: "number", Value: req.Version.Number},
		},
	}

	if err := json.NewEncoder(i.OutStream).Encode(res); err != nil {
		return i.fatal("encoding response", err)
	}

	return 0
}

func (i *In) fatal(doing string, err error) int {
	fmt.Fprintf(i.ErrStream, "error %s: %v", doing, err)
	return 1
}

func main() {
	c := &In{
		InStream:  os.Stdin,
		ErrStream: os.Stderr,
		OutStream: os.Stdout,
	}
	os.Exit(c.Execute(os.Args))
}
