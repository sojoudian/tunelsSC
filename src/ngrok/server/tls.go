package server

import (
	"crypto/tls"
	"os"

	"github.com/inconshreveable/ngrok/src/ngrok/server/assets"
)

func LoadTLSConfig(crtPath string, keyPath string) (tlsConfig *tls.Config, err error) {
	fileOrAsset := func(path string, default_path string) ([]byte, error) {
		loadFn := os.ReadFile
		if path == "" {
			loadFn = assets.Asset
			path = default_path
		}

		return loadFn(path)
	}

	var (
		crt  []byte
		key  []byte
		cert tls.Certificate
	)

	if crt, err = fileOrAsset(crtPath, "assets/server/tls/snakeoil.crt"); err != nil {
		return
	}

	if key, err = fileOrAsset(keyPath, "assets/server/tls/snakeoil.key"); err != nil {
		return
	}

	if cert, err = tls.X509KeyPair(crt, key); err != nil {
		return
	}

	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			// TLS 1.3 cipher suites
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			// TLS 1.2 cipher suites for compatibility
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
	}

	return
}
