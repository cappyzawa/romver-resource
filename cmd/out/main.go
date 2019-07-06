package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	resource "github.com/cappyzawa/romver-resource"
	"github.com/cappyzawa/romver-resource/driver"
)

// Out represents out command stream
type Out struct {
	InStream  io.Reader
	ErrStream io.Writer
	OutStream io.Writer
}

// Execute executes out command
func (o *Out) Execute(args []string) int {
	if len(args) < 2 {
		fmt.Fprintf(o.ErrStream, "usage: %s <source>", args[0])
		return 1
	}

	sourceDir := args[1]

	var req resource.OutRequest
	if err := json.NewDecoder(o.InStream).Decode(&req); err != nil {
		return o.fatal("decoding request", err)
	}

	driver, err := driver.FromSource(req.Source)
	if err != nil {
		return o.fatal("construction driver", err)
	}

	var newVersion string
	if req.Params.File != "" {
		versionFile, err := os.Open(filepath.Join(sourceDir, req.Params.File))
		if err != nil {
			return o.fatal("opening version file", err)
		}
		defer versionFile.Close()

		var versionStr string
		if _, err := fmt.Fscanf(versionFile, "%s", &versionStr); err != nil {
			return o.fatal("reading versin file", err)
		}
		newVersion = versionStr
		if err := driver.Set(newVersion); err != nil {
			return o.fatal("setting version", err)
		}
	} else if req.Params.Bump {
		newVersion, err = driver.Bump()
		if err != nil {
			return o.fatal("dumping version", err)
		}
	} else {
		fmt.Fprint(o.ErrStream, "no version dump specified")
		return 1
	}

	res := resource.OutResponse{
		Version: resource.Version{
			Number: newVersion,
		},
		Metadata: []resource.MetadataField{
			{Name: "number", Value: newVersion},
		},
	}
	if err := json.NewEncoder(o.OutStream).Encode(res); err != nil {
		return o.fatal("encoding response", err)
	}

	return 0
}

func (o *Out) fatal(doing string, err error) int {
	fmt.Fprintf(o.ErrStream, "error %s: %v", doing, err)
	return 1
}

func main() {
	c := &Out{
		InStream:  os.Stdin,
		ErrStream: os.Stderr,
		OutStream: os.Stdout,
	}
	os.Exit(c.Execute(os.Args))
}
