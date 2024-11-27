package db

import (
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Type string

const (
	PostgresDBType Type = "postgres"
	MySQLDBType    Type = "mysql"
)

// Config holds database connection config info.
type Config struct {
	Type     Type
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	// Query options
	SSLMode        string
	SSLCA          string
	SSLCert        string
	SSLKey         string
	SSLSNI         string
	ConnectTimeout uint64
}

func (db *Config) Init() {
	if len(db.Type) == 0 {
		db.Type = Type(getEnv("DATABASE_TYPE", string(PostgresDBType)))
	}
	var defPort string
	switch db.Type {
	case PostgresDBType:
		defPort = "5432"
	case MySQLDBType:
		defPort = "3306"
	}
	keyPre := strings.ToUpper(string(db.Type)) + "_"
	if len(db.Host) == 0 {
		db.Host = getEnv(keyPre+"HOST", "localhost")
	}
	if len(db.Port) == 0 {
		db.Port = getEnv(keyPre+"PORT", defPort)
	}
	if len(db.User) == 0 {
		db.User = getEnv(keyPre + "USER")
	}
	if len(db.Password) == 0 {
		db.Password = getEnv(keyPre + "PASSWORD")
	}
	if len(db.DBName) == 0 {
		db.DBName = getEnv(keyPre + "DB")
	}
	if len(db.SSLMode) == 0 {
		db.SSLMode = getEnv(keyPre + "SSLMODE")
	}
	if db.ConnectTimeout == 0 {
		db.ConnectTimeout, _ = getEnvUint(keyPre + "CONNECT_TIMEOUT")
	}
}

func (db *Config) URI() *url.URL {
	u := url.URL{
		Scheme: string(db.Type),
		Host:   net.JoinHostPort(db.Host, db.Port),
		Path:   filepath.Join("/", db.DBName),
	}
	if len(db.User) > 0 && len(db.Password) > 0 {
		u.User = url.UserPassword(db.User, db.Password)
	}
	q := make(url.Values)
	switch db.Type {
	case PostgresDBType:
		if db.ConnectTimeout > 0 {
			q.Set("connect_timeout", strconv.FormatUint(db.ConnectTimeout, 10))
		}
		if len(db.SSLMode) > 0 {
			q.Set("sslmode", db.SSLMode)
		}
		if len(db.SSLCA) > 0 {
			q.Set("sslrootcert", db.SSLCA)
		}
		if len(db.SSLCert) > 0 {
			q.Set("sslcert", db.SSLCert)
		}
		if len(db.SSLKey) > 0 {
			q.Set("sslkey", db.SSLKey)
		}
		if len(db.SSLSNI) > 0 {
			q.Set("sslsni", db.SSLSNI)
		}
	case MySQLDBType:
		if db.ConnectTimeout > 0 {
			q.Set("connect-timeout", strconv.FormatUint(db.ConnectTimeout, 10))
		}
		if len(db.SSLMode) > 0 {
			q.Set("ssl-mode", db.SSLMode)
		}
		if len(db.SSLCA) > 0 {
			q.Set("ssl-ca", db.SSLCA)
		}
		if len(db.SSLCert) > 0 {
			q.Set("ssl-cert", db.SSLCert)
		}
		if len(db.SSLKey) > 0 {
			q.Set("ssl-key", db.SSLKey)
		}
	}
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}
	return &u
}

var errEnvNotFound = errors.New("environment variable not found")

func getEnv(key string, defaults ...string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		for _, val := range defaults {
			if len(val) > 0 {
				return val
			}
		}
		return ""
	}
	return v
}

func getEnvUint(key string, defaults ...uint64) (uint64, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		if len(defaults) > 0 {
			return defaults[0], nil
		}
		return 0, errEnvNotFound
	}
	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}
