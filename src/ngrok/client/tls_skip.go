//go:build skip_verify
// +build skip_verify

package client

import (
	"crypto/tls"
)

var (
	rootCrtPaths = []string{}
)

func useInsecureSkipVerify() bool {
	return true
}

func tlsConfig(rootCAs *tls.CertPool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}
