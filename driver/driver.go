package driver

import (
	"bytes"
	"fmt"

	resource "github.com/cappyzawa/romver-resource"
)

// Driver operates the versioning
type Driver interface {
	Bump() (string, error)
	Set(string) error
	Check(string) ([]string, error)
}

// FromSource returns driver based on source configuration
func FromSource(source resource.Source) (Driver, error) {
	if source.InitialVersion == "" {
		source.InitialVersion = "0"
	}

	switch source.Driver {
	case resource.DriverUnspecified:
		return nil, fmt.Errorf("driver is empty")
	case resource.DriverGit:
		return &GitDriver{
			InitialVersion: source.InitialVersion,

			URI:           source.URI,
			Branch:        source.Branch,
			PrivateKey:    source.PrivateKey,
			Username:      source.Username,
			Password:      source.Password,
			File:          source.File,
			GitUser:       source.GitUser,
			CommitMessage: source.CommitMessage,

			Runner: &resource.ExCommand{
				Stdout: new(bytes.Buffer),
				Stderr: new(bytes.Buffer),
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown driver: %s", source.Driver)
	}
}
