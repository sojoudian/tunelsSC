package server

import (
	"crypto/tls"
	"github.com/inconshreveable/ngrok/src/ngrok/conn"
	log "github.com/inconshreveable/ngrok/src/ngrok/log"
	"github.com/inconshreveable/ngrok/src/ngrok/msg"
	"github.com/inconshreveable/ngrok/src/ngrok/ratelimit"
	"github.com/inconshreveable/ngrok/src/ngrok/util"
	"math/rand"
	"net"
	"os"
	"runtime/debug"
	"time"
)

const (
	registryCacheSize uint64        = 1024 * 1024 // 1 MB
	connReadTimeout   time.Duration = 10 * time.Second
)

// GLOBALS
var (
	tunnelRegistry  *TunnelRegistry
	controlRegistry *ControlRegistry

	// XXX: kill these global variables - they're only used in tunnel.go for constructing forwarding URLs
	opts      *Options
	listeners map[string]*conn.Listener

	// Rate limiters
	connRateLimiter *ratelimit.ConnectionRateLimiter
	ipRateLimiter   *ratelimit.IPRateLimiter
)

func NewProxy(pxyConn conn.Conn, regPxy *msg.RegProxy) {
	// fail gracefully if the proxy connection fails to register
	defer func() {
		if r := recover(); r != nil {
			pxyConn.Warn("Failed with error: %v", r)
			pxyConn.Close()
		}
	}()

	// set logging prefix
	pxyConn.SetType("pxy")

	// look up the control connection for this proxy
	pxyConn.Info("Registering new proxy for %s", regPxy.ClientId)
	ctl := controlRegistry.Get(regPxy.ClientId)

	if ctl == nil {
		panic("No client found for identifier: " + regPxy.ClientId)
	}

	ctl.RegisterProxy(pxyConn)
}

// Listen for incoming control and proxy connections
// We listen for incoming control and proxy connections on the same port
// for ease of deployment. The hope is that by running on port 443, using
// TLS and running all connections over the same port, we can bust through
// restrictive firewalls.
func tunnelListener(addr string, tlsConfig *tls.Config) {
	// listen for incoming connections
	listener, err := conn.Listen(addr, "tun", tlsConfig)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for control and proxy connections on %s", listener.Addr.String())
	for c := range listener.Conns {
		// Extract IP address for rate limiting
		remoteAddr := c.RemoteAddr().String()
		ip, _, _ := net.SplitHostPort(remoteAddr)

		// Apply rate limiting
		if ipRateLimiter != nil && !ipRateLimiter.AllowIP(ip) {
			c.Warn("Rate limit exceeded for IP: %s", ip)
			c.Close()
			continue
		}

		if connRateLimiter != nil && !connRateLimiter.AllowConnection(ip) {
			c.Warn("Connection limit exceeded for IP: %s", ip)
			c.Close()
			continue
		}

		go func(tunnelConn conn.Conn, clientIP string) {
			// don't crash on panics
			defer func() {
				if r := recover(); r != nil {
					tunnelConn.Info("tunnelListener failed with error %v: %s", r, debug.Stack())
				}
				// Release connection count
				if connRateLimiter != nil {
					connRateLimiter.ReleaseConnection(clientIP)
				}
			}()

			tunnelConn.SetReadDeadline(time.Now().Add(connReadTimeout))
			var rawMsg msg.Message
			if rawMsg, err = msg.ReadMsg(tunnelConn); err != nil {
				tunnelConn.Warn("Failed to read message: %v", err)
				tunnelConn.Close()
				return
			}

			// don't timeout after the initial read, tunnel heartbeating will kill
			// dead connections
			tunnelConn.SetReadDeadline(time.Time{})

			switch m := rawMsg.(type) {
			case *msg.Auth:
				NewControl(tunnelConn, m)

			case *msg.RegProxy:
				NewProxy(tunnelConn, m)

			default:
				tunnelConn.Close()
			}
		}(c, ip)
	}
}

func Main() {
	// parse options
	opts = parseArgs()

	// init logging
	log.LogTo(opts.logto, opts.loglevel)

	// seed random number generator
	seed, err := util.RandomSeed()
	if err != nil {
		panic(err)
	}
	rand.Seed(seed)

	// init tunnel/control registry
	registryCacheFile := os.Getenv("REGISTRY_CACHE_FILE")
	tunnelRegistry = NewTunnelRegistry(registryCacheSize, registryCacheFile)
	controlRegistry = NewControlRegistry()

	// initialize rate limiters
	// 10 connections per second per IP, burst of 20
	ipRateLimiter = ratelimit.NewIPRateLimiter(10, 20)
	// 100 connections per second globally, burst of 200, max 50 concurrent per IP
	connRateLimiter = ratelimit.NewConnectionRateLimiter(100, 200, 50)

	// start listeners
	listeners = make(map[string]*conn.Listener)

	// load tls configuration
	tlsConfig, err := LoadTLSConfig(opts.tlsCrt, opts.tlsKey)
	if err != nil {
		panic(err)
	}

	// listen for http
	if opts.httpAddr != "" {
		listeners["http"] = startHttpListener(opts.httpAddr, nil)
	}

	// listen for https
	if opts.httpsAddr != "" {
		listeners["https"] = startHttpListener(opts.httpsAddr, tlsConfig)
	}

	// ngrok clients
	tunnelListener(opts.tunnelAddr, tlsConfig)
}
