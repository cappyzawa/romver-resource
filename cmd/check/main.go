package main

import (
	"encoding/json"
	"fmt"
	"os"

	resource "github.com/cappyzawa/romver-resource"
)

func main() {
	var req resource.CheckRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode: %v\n", err)
		os.Exit(1)
	}
	var res resource.CheckResponse
	if req.Version != nil {
		res = append(res, *req.Version)
	}
	if err := json.NewEncoder(os.Stdout).Encode(res); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
