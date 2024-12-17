package cloudapi

import "time"

type Config struct {

	// read at https://cloud.kt.com/docs/open-api-guide/d/guide/how-to-use

	// base url for KT cloud apis
	ApiBaseURL string `envconfig:"API_BASE_URL" default:"https://api.ucloudbiz.olleh.com/"`

	// api zone
	Zone string `envconfig:"ZONE" default:"gd1"`

	// Use "password" value by default
	IdentityMethods string `envconfig:"IDENTITY_METHODS" default:"password"`

	// Use "default" value by default
	IdentityPasswordUserDomainId string `envconfig:"IDENTITY_PASSWORD_USER_DOMAIN_ID" default:"default"`

	// User's id
	IdentityPasswordUserName string `envconfig:"IDENTITY_PASSWORD_USERNAME" default:"soongsil_a050_gov@vple.net"`

	// User's password
	IdentityPassword string `envconfig:"IDENTITY_PASSWORD" default:"Qlenfrl!#24"`

	// Use "default" value by default
	ScopeProjectDomainId string `envconfig:"SCOPE_PROJECT_DOMAIN_ID" default:"default"`

	// User's id
	ScopeProjectName string `envconfig:"SCOPE_PROJECT_NAME" default:"soongsil_a050_gov@vple.net"`

	// Verbosity of the logger.
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// Client token handling approach
	ClientAuthAutoRenew bool `envconfig:"CLIENT_AUTH_AUTO_RENEW" default:"true"`

	// RemoteBackendTimeout specifies timeout. Has to be parsable to time.Duration
	RemoteBackendTimeout time.Duration `envconfig:"REMOTE_BACKEND_TIMEOUT" default:"5s"`
}
