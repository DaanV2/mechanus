// Package webclient holds the compiled web client embedded in the server binary, and installs it into
// the static folder the server serves.
//
// The embedded files come from dist/, which `go run ./scripts/embedclient` fills with client/build
// before a release build. A binary built without that step embeds nothing and serves the static folder
// as it is.
package webclient
