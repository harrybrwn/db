package db

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestConfig_Init(t *testing.T) {
	type KV struct{ K, V string }
	type table struct {
		env []KV
		exp Config
		uri string
	}
	is := is.New(t)

	for _, tt := range []table{
		{
			env: []KV{
				{"DATABASE_TYPE", "postgres"},
				{"POSTGRES_HOST", "127.0.0.2"},
				{"POSTGRES_PORT", "1234"},
				{"POSTGRES_DB", "test_db"},
				{"POSTGRES_CONNECT_TIMEOUT", "69"},
				{"POSTGRES_SSLMODE", "any_dumb_value"},
			},
			exp: Config{
				Type:           PostgresDBType,
				Host:           "127.0.0.2",
				Port:           "1234",
				DBName:         "test_db",
				ConnectTimeout: 69,
				SSLMode:        "any_dumb_value",
			},
			uri: "postgres://127.0.0.2:1234/test_db?connect_timeout=69&sslmode=any_dumb_value",
		},
		{
			env: []KV{
				{"DATABASE_TYPE", "mysql"},
				{"MYSQL_HOST", "127.0.0.3"},
				{"MYSQL_PORT", "5555"},
				{"MYSQL_DB", "jimmyjohns"},
				{"MYSQL_CONNECT_TIMEOUT", "123"},
				{"MYSQL_SSLMODE", "off"},
				// should be ignored
				{"POSTGRES_HOST", "127.0.0.2"},
				{"POSTGRES_PORT", "1234"},
				{"POSTGRES_DB", "test_db"},
				{"POSTGRES_CONNECT_TIMEOUT", "69"},
				{"POSTGRES_SSLMODE", "any_dumb_value"},
			},
			exp: Config{
				Type:           MySQLDBType,
				Host:           "127.0.0.3",
				Port:           "5555",
				DBName:         "jimmyjohns",
				SSLMode:        "off",
				ConnectTimeout: 123,
			},
			uri: "mysql://127.0.0.3:5555/jimmyjohns?connect-timeout=123&ssl-mode=off",
		},
	} {
		clearEnv()
		for _, e := range tt.env {
			os.Setenv(e.K, e.V)
		}
		var c Config
		c.Init()
		is.Equal(c, tt.exp)
		is.Equal(c.URI().String(), tt.uri)
	}
}

func TestConfig_URI(t *testing.T) {
	is := is.New(t)
	var d Config

	clearEnv()
	d.Init()
	is.Equal(d.URI().String(), "postgres://localhost:5432/")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "password1")
	d.Init()
	is.Equal(d.URI().String(), "postgres://testuser:password1@localhost:5432/")
	d.DBName = "db"
	is.Equal(d.URI().String(), "postgres://testuser:password1@localhost:5432/db")
	d.SSLMode = "disable"
	d.ConnectTimeout = 9
	is.Equal(d.URI().String(), "postgres://testuser:password1@localhost:5432/db?connect_timeout=9&sslmode=disable")
	d.SSLCA = "ca.crt"
	d.SSLCert = "ssl.crt"
	d.SSLKey = "ssl.key"
	d.SSLSNI = "sni"
	is.Equal(d.URI().String(), "postgres://testuser:password1@localhost:5432/db?connect_timeout=9&sslcert=ssl.crt&sslkey=ssl.key&sslmode=disable&sslrootcert=ca.crt&sslsni=sni")

	d.Type = MySQLDBType
	d.Port = ""
	d.Init()
	is.Equal(d.URI().String(), "mysql://testuser:password1@localhost:3306/db?connect-timeout=9&ssl-ca=ca.crt&ssl-cert=ssl.crt&ssl-key=ssl.key&ssl-mode=disable")
	d.ConnectTimeout = 0
	is.Equal(d.URI().String(), "mysql://testuser:password1@localhost:3306/db?ssl-ca=ca.crt&ssl-cert=ssl.crt&ssl-key=ssl.key&ssl-mode=disable")
	os.Setenv("MYSQL_CONNECT_TIMEOUT", "3")
	d.Init()
	is.Equal(d.URI().String(), "mysql://testuser:password1@localhost:3306/db?connect-timeout=3&ssl-ca=ca.crt&ssl-cert=ssl.crt&ssl-key=ssl.key&ssl-mode=disable")
}

func TestUtils(t *testing.T) {
	is := is.New(t)
	v, err := getEnvUint("__NOT_HERE__", 25)
	is.NoErr(err)
	is.Equal(v, uint64(25))
	os.Setenv("SOME_VALUE", "not-a-number")
	_, err = getEnvUint("SOME_VALUE")
	is.True(errors.Is(err, strconv.ErrSyntax))
	syntaxErr := err.(*strconv.NumError)
	is.Equal(syntaxErr.Func, "ParseUint")
	is.Equal(syntaxErr.Num, "not-a-number")
	is.Equal(syntaxErr.Err, strconv.ErrSyntax)
}

func clearEnv() {
	os.Unsetenv("DATABASE_TYPE")
	for _, tp := range []Type{PostgresDBType, MySQLDBType} {
		t := strings.ToUpper(string(tp))
		os.Unsetenv(t + "_HOST")
		os.Unsetenv(t + "_PORT")
		os.Unsetenv(t + "_USER")
		os.Unsetenv(t + "_PASSWORD")
		os.Unsetenv(t + "_DB")
		os.Unsetenv(t + "_SSLMODE")
		os.Unsetenv(t + "_CONNECT_TIMEOUT")
	}
}
