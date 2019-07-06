package resource

// Driver represents the driver
// now, git only
type Driver string

const (
	// DriverUnspecified is empty
	DriverUnspecified Driver = ""
	// DriverGit for git
	DriverGit Driver = "git"
)

// CheckRequest represents the request for checking resource
type CheckRequest struct {
	Source  Source   `json:"source"`
	Version *Version `json:"version"`
}

// CheckResponse represents the response of checking resorce
type CheckResponse []Version

// InRequest represents the request for get step
type InRequest struct {
	Source  Source   `json:"source"`
	Version Version  `json:"version"`
	Params  InParams `json:"params"`
}

// InParams represents the parameters for get step
type InParams struct {
	Bump bool `json:"bump"`
}

// InResponse represents the response of get step
type InResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

// OutRequest represents the request for put step
type OutRequest struct {
	Source Source    `json:"source"`
	Params OutParams `json:"params"`
}

// OutParams represents the parameters for put step
type OutParams struct {
	File string `json:"file"`
	Bump bool   `json:"bump"`
}

// OutResponse represents the response of put step
type OutResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

// Source represents the source configuration
type Source struct {
	Driver Driver `json:"driver"`

	InitialVersion string `json:"initial_version"`

	URI           string `json:"uri"`
	Branch        string `json:"branch"`
	PrivateKey    string `json:"private_key"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	File          string `json:"file"`
	GitUser       string `json:"git_user"`
	CommitMessage string `json:"commit_message"`
}

// Version represents the resource version
type Version struct {
	Number string `json:"number"`
}

// Metadata represents the resorce metadata
type Metadata []MetadataField

// MetadataField represents key/value of metadata
type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
