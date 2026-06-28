package bootstrap

import (
	"go.uber.org/zap"

	governhttp "github.com/haipham22/govern/http"

	"github.com/haipham22/golang-sample/internal/handler/rest"
	"github.com/haipham22/golang-sample/pkg/config"
)

// NewHTTPServer delegates to internal/handler/rest.New, which owns the full
// HTTP wiring (db -> repository -> service -> controllers -> echo -> server).
// Keeping the wiring in rest/ preserves the existing integration tests; this
// function is the seam that lets bootstrap assemble the App without
// duplicating the dependency graph.
func NewHTTPServer(
	log *zap.SugaredLogger,
	port int64,
	appConfig *config.EnvConfigMap,
) (governhttp.Server, func(), error) {
	return rest.New(log, port, appConfig)
}
