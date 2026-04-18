package server

// Container healthcheck subcommand. Distroless runtime images (static +
// base) ship without /bin/sh / wget / curl, so a compose `CMD-SHELL`
// healthcheck can't run. Calling the service binary itself with a flag
// lets the image probe its own /health endpoint without shell support.
//
// Wire-up at the top of every service main.go:
//
//	func main() {
//	    server.RunHealthcheckAndExit("APP_PORT", "8020")
//	    ...
//	}
//
// And in docker-compose:
//
//	healthcheck:
//	  test: ["CMD", "/app", "--healthcheck"]

import (
	"net/http"
	"os"
	"time"

	"github.com/Camionerou/rag-saldivia/pkg/config"
)

// healthcheckClient has a short timeout so a stuck service doesn't keep
// the container in the starting state indefinitely — compose's healthcheck
// has its own timeout but a dangling probe still ties up goroutines.
var healthcheckClient = &http.Client{Timeout: 3 * time.Second}

// RunHealthcheckAndExit, when invoked with `--healthcheck` as the first
// CLI arg, probes http://localhost:<port>/health and terminates the
// process with exit 0 on HTTP 2xx, 1 otherwise. Returns without doing
// anything if the flag isn't set, so a no-arg invocation boots normally.
//
// Call this as the first line of main() — before server.New() — so the
// probe runs without starting the full service stack.
func RunHealthcheckAndExit(portEnvVar, defaultPort string) {
	if len(os.Args) < 2 || os.Args[1] != "--healthcheck" {
		return
	}
	port := config.Env(portEnvVar, defaultPort)
	os.Exit(healthcheckStatus(port))
}

// healthcheckStatus is factored out of RunHealthcheckAndExit so tests can
// exercise the HTTP behaviour without invoking os.Exit.
func healthcheckStatus(port string) int {
	resp, err := healthcheckClient.Get("http://localhost:" + port + "/health")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return 0
	}
	return 1
}
