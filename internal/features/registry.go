package features

import (
	authfeature "github.com/naodEthiop/lalibela-cli/internal/features/auth"
	configfeature "github.com/naodEthiop/lalibela-cli/internal/features/config"
	corsfeature "github.com/naodEthiop/lalibela-cli/internal/features/cors"
	dockerfeature "github.com/naodEthiop/lalibela-cli/internal/features/docker"
	errorhandlerfeature "github.com/naodEthiop/lalibela-cli/internal/features/errorhandler"
	gracefulshutdownfeature "github.com/naodEthiop/lalibela-cli/internal/features/gracefulshutdown"
	healthfeature "github.com/naodEthiop/lalibela-cli/internal/features/health"
	loggerfeature "github.com/naodEthiop/lalibela-cli/internal/features/logger"
	postgresfeature "github.com/naodEthiop/lalibela-cli/internal/features/postgres"
	ratelimitfeature "github.com/naodEthiop/lalibela-cli/internal/features/ratelimit"
	redisfeature "github.com/naodEthiop/lalibela-cli/internal/features/redis"
	swaggerfeature "github.com/naodEthiop/lalibela-cli/internal/features/swagger"
)

var Registry = map[string]Feature{
	"auth":              authfeature.New(),
	"config":            configfeature.New(),
	"cors":              corsfeature.New(),
	"docker":            dockerfeature.New(),
	"error-handler":     errorhandlerfeature.New(),
	"graceful-shutdown": gracefulshutdownfeature.New(),
	"health":            healthfeature.New(),
	"logger":            loggerfeature.New(),
	"postgres":          postgresfeature.New(),
	"rate-limit":        ratelimitfeature.New(),
	"redis":             redisfeature.New(),
	"swagger":           swaggerfeature.New(),
}

var DefaultProductionFeatures = []string{
	"config",
	"logger",
	"graceful-shutdown",
	"health",
	"error-handler",
	"cors",
}
