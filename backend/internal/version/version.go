// Package version exposes build identity for the Atlas API process.
package version

// Name is the service name returned by the /version endpoint.
const Name = "atlas"

// Version is the application version string for this release.
const Version = "0.1.0"

// Info is the JSON payload for GET /version.
type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Current returns the compiled-in service identity.
func Current() Info {
	return Info{Name: Name, Version: Version}
}
