package services

import (
	"github.com/boeing-ai-gateway/kinm/pkg/db"
	"github.com/boeing-ai-gateway/nah/pkg/randomtoken"
	"github.com/boeing-ai-gateway/boeing/logger"
	"github.com/boeing-ai-gateway/boeing/pkg/logutil"
	"github.com/boeing-ai-gateway/boeing/pkg/storage/authn"
	"github.com/boeing-ai-gateway/boeing/pkg/storage/authz"
	"github.com/boeing-ai-gateway/boeing/pkg/storage/scheme"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

var log = logger.Package()

type Config struct {
	StorageListenPort int    `usage:"Port to storage backend will listen on (default: random port)"`
	StorageToken      string `usage:"Token for storage access, will be generated if not passed"`
	DSN               string `usage:"Database dsn in driver://connection_string format" default:"sqlite://file:boeing.db?_journal=WAL&cache=shared&_busy_timeout=30000"`
}

type Services struct {
	DB    *db.Factory
	Authn *authn.Authenticator
	Authz authorizer.Authorizer
}

func New(config Config) (_ *Services, err error) {
	if config.StorageToken == "" {
		config.StorageToken, err = randomtoken.Generate()
		if err != nil {
			return nil, err
		}
	}

	// Sanitize DSN for logging (remove credentials)
	sanitizedDSN := logutil.SanitizeDSN(config.DSN)
	log.Debugf("Creating database factory. dsn: %v", sanitizedDSN)
	dbClient, err := db.NewFactory(scheme.Scheme, config.DSN)
	if err != nil {
		log.Errorf("Failed to create database factory: dsn=%s error=%v", sanitizedDSN, err)
		return nil, err
	}
	log.Debugf("Database factory created successfully. dsn: %v", sanitizedDSN)

	services := &Services{
		DB:    dbClient,
		Authn: authn.NewAuthenticator(config.StorageToken),
		Authz: &authz.Authorizer{},
	}

	return services, nil
}
