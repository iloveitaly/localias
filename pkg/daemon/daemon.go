package daemon

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	godaemon "github.com/sevlyar/go-daemon"

	"github.com/peterldowns/localias/pkg/config"
	"github.com/peterldowns/localias/pkg/hostctl"
)

// Run will apply the latest configuration and start the caddy server, blocking
// indefinitely until the process is terminated.
func Run(hctl *hostctl.Controller, cfg *config.Config) error {
	err := applyCfg(hctl, cfg)
	if err != nil {
		return err
	}
	cfgJSON, _, err := cfg.CaddyJSON()
	if err != nil {
		return err
	}
	err = caddy.Load(cfgJSON, false)
	if err != nil {
		return err
	}
	select {} //nolint:revive // valid empty block, keeps the server running forever.
}

// Start will apply the latest configuration and start the caddy daemon server,
// then exit. If the caddy daemon server is already running, it will exit with
// an error.
func Start(hctl *hostctl.Controller, cfg *config.Config) error {
	existing, err := Status()
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("daemon is already running")
	}

	cntxt := daemonContext()
	d, err := cntxt.Reborn()
	if err != nil {
		return err
	}
	// parent process: the child has started, so exit
	if d != nil {
		return nil
	}
	// child process: defer a cleanup function, then run
	defer func() {
		_ = cntxt.Release()
	}()
	return Run(hctl, cfg)
}

// Status will determine whether or not the caddy daemon server is running.  If
// it is, it returns the non-nil os.Process of that daemon.
func Status() (*os.Process, error) {
	cntxt := daemonContext()
	return cntxt.Search()
}

// Stop will attempt to stop the caddy daemon server by sending an API request
// over http. If the daemon server is not running, it will return an error.
func Stop(cfg *config.Config) error {
	address, err := determineAPIAddress(cfg)
	if err != nil {
		return fmt.Errorf("could not determine api address: %w", err)
	}
	resp, err := caddycmd.AdminAPIRequest(address, http.MethodPost, "/stop", nil, nil)
	if err != nil {
		return fmt.Errorf("request to /stop failed: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// Reload will apply the latest configuration details, and then update the
// running caddy daemon server's configuration by sending an API request over
// http.  If the daemon server is not running, it will return an error.
func Reload(hctl *hostctl.Controller, cfg *config.Config) error {
	err := applyCfg(hctl, cfg)
	if err != nil {
		return err
	}
	cfgJSON, _, err := cfg.CaddyJSON()
	if err != nil {
		return err
	}
	address, err := determineAPIAddress(cfg)
	if err != nil {
		return err
	}
	headers := make(http.Header)
	headers.Set("Cache-Control", "must-revalidate")
	resp, err := caddycmd.AdminAPIRequest(address, http.MethodPost, "/load", headers, bytes.NewReader(cfgJSON))
	if err != nil {
		return fmt.Errorf("failed to send config to daemon: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// applyCfg will apply the latest configuration using hctl.
func applyCfg(hctl *hostctl.Controller, cfg *config.Config) error {
	if err := hctl.Clear(); err != nil {
		return err
	}
	for _, directive := range cfg.Directives {
		up, err := httpcaddyfile.ParseAddress(directive.Upstream)
		if err != nil {
			return err
		}
		if err := hctl.Set("127.0.0.1", up.Host); err != nil {
			return err
		}
	}
	return hctl.Apply()
}

// daemonContext returns a consistent godaemon context that is used to control
// the caddy daemon server.
func daemonContext() *godaemon.Context {
	return &godaemon.Context{
		PidFileName: "localias.pid",
		PidFilePerm: 0o644,
		WorkDir:     "./",
		Umask:       0o27,
	}
}
