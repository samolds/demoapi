package config

import (
	"flag"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"github.com/hashicorp/hcl"
	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"
)

var (
	configFileFlag = flag.String("config", "", "config file path")
	dbURLFlag      = flag.String("db_url", "", "database url") // flag override
	logLevelFlag   = flag.String("loglevel", "", "log level")  // flag override

	// env var overrides
	dbDriverEnv = os.Getenv("DATABASE_DRIVER")
	dbHostEnv   = os.Getenv("DATABASE_HOST")
	dbPortEnv   = os.Getenv("DATABASE_PORT")
	dbSSLEnv    = os.Getenv("DATABASE_SSLMODE")
	dbUserEnv   = os.Getenv("DATABASE_USER")
	dbPassEnv   = os.Getenv("DATABASE_PASSWORD")
	dbDBNameEnv = os.Getenv("DATABASE_NAME")

	// handled explicitly because the postgres server expects these env vars
	psqlUserEnv   = os.Getenv("POSTGRES_USER")
	psqlPassEnv   = os.Getenv("POSTGRES_PASSWORD")
	psqlDBNameEnv = os.Getenv("POSTGRES_DB")

	configErr = errs.Class("configuration")
)

type Configs struct {
	DBURL                   *url.URL
	APISlug                 string
	APIAddress              string
	MetricAddress           string
	GracefulShutdownTimeout time.Duration
	WriteTimeout            time.Duration
	ReadTimeout             time.Duration
	IdleTimeout             time.Duration
	LogLevel                logrus.Level
	DeveloperMode           bool
	InsecureRequestsMode    bool
}

// Parse will set the configuration values pulled from the provided config
// file, any config flags, and any environment variables. env vars have highest
// priority, then config flags, then the config file. If any values overlap and
// *DON'T* match, an error is thrown.
func Parse() (*Configs, error) {
	flag.Parse()

	raw := rawConfigs{}
	err := raw.setConfigFile()
	if err != nil {
		return nil, err
	}

	err = raw.setNoChangeConfigFlags()
	if err != nil {
		return nil, err
	}

	err = raw.setNoChangeEnvVars()
	if err != nil {
		return nil, err
	}

	return raw.validate()
}

type rawConfigs struct {
	DBURL                   string `hcl:"db_url"`
	APISlug                 string `hcl:"api_slug"`
	APIAddress              string `hcl:"api_addr"`
	MetricAddress           string `hcl:"metric_addr"`
	GracefulShutdownTimeout int    `hcl:"graceful_shutdown_timeout_sec"`
	WriteTimeout            int    `hcl:"write_timeout_sec"`
	ReadTimeout             int    `hcl:"read_timeout_sec"`
	IdleTimeout             int    `hcl:"idle_timeout_sec"`
	LogLevel                string `hcl:"loglevel"`
	DeveloperMode           bool   `hcl:"developer_mode"`
	InsecureRequestsMode    bool   `hcl:"insecure_requests_mode"`
}

// setConfigFile will set all of the values provided in the config file,
// stomping over any existing values
func (raw *rawConfigs) setConfigFile() error {
	if configFileFlag == nil || *configFileFlag == "" {
		return nil
	}

	hclBytes, err := ioutil.ReadFile(*configFileFlag)
	if err != nil {
		return configErr.Wrap(err)
	}

	if err := hcl.Unmarshal(hclBytes, raw); err != nil {
		return configErr.Wrap(err)
	}

	return nil
}

// setNoChangeConfigFlags will set all of the values provided as config flags
// as long as they don't change existing non-empty values in Configs
func (raw *rawConfigs) setNoChangeConfigFlags() error {
	if dbURLFlag != nil && *dbURLFlag != "" {
		if raw.DBURL == "" {
			raw.DBURL = *dbURLFlag
		} else if raw.DBURL != *dbURLFlag {
			return configErr.New("db urls %q and %q don't match", raw.DBURL,
				*dbURLFlag)
		}
	}

	if logLevelFlag != nil && *logLevelFlag != "" {
		if raw.LogLevel == "" {
			raw.LogLevel = *logLevelFlag
		} else if raw.LogLevel != *logLevelFlag {
			return configErr.New("log level %q and %q don't match", raw.LogLevel,
				*logLevelFlag)
		}
	}

	return nil
}

// setNoChangeEnvVars will set all of the values provided as environment vars
// as long as they don't change existing non-empty values in Configs
func (raw *rawConfigs) setNoChangeEnvVars() error {

	// use dbUserEnv and psqlUserEnv as sentinel values. if these env vars
	// are set, assume the other ones exist.
	if psqlUserEnv == "" || dbUserEnv == "" {
		if dbDriverEnv != "" || dbHostEnv != "" || dbPortEnv != "" ||
			dbSSLEnv != "" || dbUserEnv != "" || dbPassEnv != "" ||
			dbDBNameEnv != "" || psqlUserEnv != "" || psqlPassEnv != "" ||
			psqlDBNameEnv != "" {
			return configErr.New("database connection env vars incompletely set")
		}

		return nil
	}

	if dbDriverEnv == "postgres" {
		if psqlUserEnv == "" || psqlPassEnv == "" || psqlDBNameEnv == "" {
			return configErr.New("required postgres env vars missing")
		}

		if (dbUserEnv != "" || dbPassEnv != "" || dbDBNameEnv != "") &&
			(dbUserEnv != psqlUserEnv || dbPassEnv != psqlPassEnv ||
				dbDBNameEnv != psqlDBNameEnv) {
			return configErr.New("non matching db and psql specific env vars set")
		}
	}

	envVarDB := &url.URL{
		Scheme:   dbDriverEnv,
		User:     url.UserPassword(dbUserEnv, dbPassEnv),
		Host:     dbHostEnv + ":" + dbPortEnv,
		Path:     dbDBNameEnv,
		RawQuery: "sslmode=" + dbSSLEnv,
	}

	envVarDBString := envVarDB.String()
	if raw.DBURL == "" {
		raw.DBURL = envVarDBString
	} else if raw.DBURL != envVarDBString {
		return configErr.New("db urls %q and env %q don't match", raw.DBURL,
			envVarDBString)
	}

	return nil
}

func (raw *rawConfigs) validate() (*Configs, error) {
	if raw.DBURL == "" {
		return nil, configErr.New("db_url misconfigured")
	}
	if raw.APISlug == "" {
		return nil, configErr.New("api_slug misconfigured")
	}
	if raw.APIAddress == "" {
		return nil, configErr.New("api_addr misconfigured")
	}
	if raw.MetricAddress == "" {
		return nil, configErr.New("metric_addr misconfigured")
	}
	if raw.GracefulShutdownTimeout == 0 {
		return nil, configErr.New("graceful_shutdown_timeout_sec misconfigured")
	}
	if raw.WriteTimeout == 0 {
		return nil, configErr.New("write_sec misconfigured")
	}
	if raw.ReadTimeout == 0 {
		return nil, configErr.New("read_sec misconfigured")
	}
	if raw.IdleTimeout == 0 {
		return nil, configErr.New("idle_sec misconfigured")
	}
	if raw.LogLevel == "" {
		return nil, configErr.New("loglevel misconfigured")
	}

	dbURL, err := url.Parse(raw.DBURL)
	if err != nil {
		return nil, err
	}

	grace := time.Second * time.Duration(raw.GracefulShutdownTimeout)
	write := time.Second * time.Duration(raw.WriteTimeout)
	read := time.Second * time.Duration(raw.ReadTimeout)
	idle := time.Second * time.Duration(raw.IdleTimeout)

	loglevel, err := logrus.ParseLevel(raw.LogLevel)
	if err != nil {
		return nil, err
	}

	return &Configs{
		DBURL:                   dbURL,
		APISlug:                 raw.APISlug,
		APIAddress:              raw.APIAddress,
		MetricAddress:           raw.MetricAddress,
		GracefulShutdownTimeout: grace,
		WriteTimeout:            write,
		ReadTimeout:             read,
		IdleTimeout:             idle,
		LogLevel:                loglevel,
		DeveloperMode:           raw.DeveloperMode,
		InsecureRequestsMode:    raw.InsecureRequestsMode,
	}, nil
}
