package database

import (
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"

	monitor "demoapi/prometheus"
)

const (
	SqliteDriver   = "sqlite3"
	PostgresDriver = "postgres"
)

var (
	dbErr = errs.Class("database")
)

type Database struct {
	*DB
	driver string
}

type Config struct {
	MaxOpenConns *int
	MaxIdleConns *int
}

// TODO(sam): this database package needs a lot of love. there should be a
// database interface to make supporting multiple database drivers easier and
// cleaner. all of the switches are gross.
// TODO(sam): add support for migrations! Critical importance.
func Connect(dbURL *url.URL, c *Config) (*Database, error) {
	// WrapErr is a dbx specific error wrapping hook
	WrapErr = StacktraceWrapAnyError

	// Logger is a dbx specific logging hook. it's called by every dbx query
	Logger = func(format string, args ...interface{}) {
		logrus.Debugf(format, args...)

		go func() {
			// it is perhaps an abuse of this logger to collect metrics with it
			monitor.DatabaseQueryCounter.Inc()
		}()
	}

	// copy the dbURL and remove user/password so it's loggable
	loggableURL, err := url.Parse(dbURL.String())
	if err != nil {
		return nil, err
	}
	loggableURL.User = nil // don't log username/password
	logrus.Infof("connecting to db: %s", loggableURL.String())

	driver := strings.ToLower(dbURL.Scheme)
	if driver == "sqlite3" {
		dbURL.Scheme = "file"
	}

	dbConn, err := Open(driver, dbURL.String())
	if err != nil {
		return nil, err
	}

	db, err := newDatabase(driver, dbConn)
	if err != nil {
		return nil, err
	}

	if db.brandNew() {
		logrus.Debugf("%s will be initialized", loggableURL.String())
		err = db.initializeSchema()
		if err != nil {
			return nil, err
		}
	} else {
		logrus.Debugf("%s already exists", loggableURL.String())
	}

	db.configure(c)
	logrus.Infof("connected to database")

	// TODO(sam): add support for migrations!
	// err = db.migrate()
	// if err != nil {
	//   return nil, err
	// }

	// TODO(sam): spin off a lightweight goroutine that will occasionally query
	// the database for counts of various tables and Set monitor metrics

	return db, nil
}

func newDatabase(driver string, dbConn *DB) (*Database, error) {
	db := &Database{DB: dbConn, driver: driver}
	return db, nil
}

func (db *Database) brandNew() bool {
	// use the users table as a sentinel of existence
	_, err := db.DB.Exec("SELECT * FROM users LIMIT 1")
	return err != nil
}

func (db *Database) initializeSchema() error {
	_, err := db.DB.Exec(db.DB.Schema())
	if err != nil {
		return dbErr.Wrap(err)
	}

	switch db.driver {
	case PostgresDriver:
	case SqliteDriver:
		_, err = db.DB.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			return dbErr.Wrap(err)
		}
	default:
		return dbErr.New("unsupported driver %q", db.driver)
	}

	return nil
}

func (db *Database) configure(c *Config) {
	if c == nil {
		return
	}

	if c.MaxOpenConns != nil {
		db.DB.SetMaxOpenConns(*c.MaxOpenConns)
	}
	if c.MaxIdleConns != nil {
		db.DB.SetMaxIdleConns(*c.MaxIdleConns)
	}
}
