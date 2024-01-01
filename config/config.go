package config

import (
	"os"
	"time"

	"github.com/JeremyLoy/config"
)

const (
	// TemplateExt stores the extension used for the template files
	TemplateExt = ".gohtml"

	// StaticDir stores the name of the directory that will serve static files
	StaticDir = "static"

	// StaticPrefix stores the URL prefix used when serving static files
	StaticPrefix = "files"
)

type environment string

const (
	// EnvLocal represents the local environment
	EnvLocal environment = "local"

	// EnvTest represents the test environment
	EnvTest environment = "test"

	// EnvDevelop represents the development environment
	EnvDevelop environment = "dev"

	// EnvStaging represents the staging environment
	EnvStaging environment = "staging"

	// EnvQA represents the qa environment
	EnvQA environment = "qa"

	// EnvProduction represents the production environment
	EnvProduction environment = "prod"
)

// SwitchEnvironment sets the environment variable used to dictate which environment the application is
// currently running in.
// This must be called prior to loading the configuration in order for it to take effect.
func SwitchEnvironment(env environment) {
	if err := os.Setenv("PAGODA_APP_ENVIRONMENT", string(env)); err != nil {
		panic(err)
	}
}

type Config struct {
	configFile string

	// HTTPConfig stores HTTP configuration
	HTTP struct {
		Hostname     string `config:"HOSTNAME"`
		Port         uint16 `config:"PORT"`
		ReadTimeout  time.Duration `config:"READ_TIMEOUT"`
		WriteTimeout time.Duration `config:"WRITE_TIMEOUT"`
		IdleTimeout  time.Duration `config:"IDLE_TIMEOUT"`
		TLS          struct {
			Enabled     bool `config:"ENABLED"`
			Certificate string `config:"CERTIFICATE"`
			Key         string `config:"KEY"`
		} `config:"TLS"`
	}

	// AppConfig stores application configuration
	App struct {
		Name          string `config:"NAME"`
		Environment   environment `config:"ENVIRONMENT"`
		EncryptionKey string `config:"ENCRYPTION_KEY"`
		Timeout       time.Duration `config:"TIMEOUT"`
		PasswordToken struct {
			Expiration time.Duration `config:"EXPIRATION"`
			Length     int `config:"LENGTH"`
		} `config:"PASSWORD_TOKEN"`
		EmailVerificationTokenExpiration time.Duration `config:"EMAIL_VERIFICATION_TOKEN_EXPIRATION"`
	}

	// CacheConfig stores the cache configuration
	Cache struct {
		Hostname     string `config:"HOSTNAME"`
		Port         uint16 `config:"PORT"`
		Password     string `config:"PASSWORD"`
		Database     int `config:"DATABASE"`
		TestDatabase int `config:"TEST_DATABASE"`
		Expiration   struct {
			StaticFile time.Duration `config:"STATIC_FILE"`
			Page       time.Duration `config:"PAGE"`
		} `config:"EXPIRATION"`
	}

	// DatabaseConfig stores the database configuration
	Database struct {
		Hostname     string `config:"HOSTNAME"`
		Port         uint16 `config:"PORT"`
		User         string `config:"USER"`
		Password     string `config:"PASSWORD"`
		Database     string `config:"DATABASE"`
		TestDatabase string `config:"TEST_DATABASE"`
	}

	// MailConfig stores the mail configuration
	Mail struct {
		Hostname    string `config:"HOSTNAME"`
		Port        uint16 `config:"PORT"`
		User        string `config:"USER"`
		Password    string `config:"PASSWORD"`
		FromAddress string `config:"FROM_ADDRESS"`
	}

	Google struct {
		ClientID		string `config:"CLIENT_ID"`
		ClientSecret    string `config:"CLIENT_SECRET"`
		RedirectURL     string `config:"REDIRECT_URL"`
	}
}
var conf = Config{
}

// GetConfig loads and returns configuration
func GetConfig() (Config, error) {
	return conf, nil
}

func init() {
	configFile := ".env"
	if f := os.Getenv("CONFIG_FILE"); f != "" {
		configFile = f
	}
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		conf.configFile = configFile
		config.From(configFile).FromEnv().To(&conf)
	} else {
		config.FromEnv().To(&conf)
	}
}

func (c Config) ConfigFile() string {
	return c.configFile
}
