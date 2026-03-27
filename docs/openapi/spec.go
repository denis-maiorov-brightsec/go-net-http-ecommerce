package openapi

import _ "embed"

var (
	//go:embed openapi.json
	spec []byte
	//go:embed ui.html
	ui []byte
)

func Spec() []byte {
	return append([]byte(nil), spec...)
}

func UI() []byte {
	return append([]byte(nil), ui...)
}
