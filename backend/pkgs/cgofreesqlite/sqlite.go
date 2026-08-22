// Package cgofreesqlite package provides a CGO free implementation of the sqlite3 driver. This wraps the
// modernc.org/sqlite driver and adds the PRAGMA foreign_keys = ON; statement to the connection
// initialization as well as registering the driver with the sql package as "sqlite3" for compatibility
// with entgo.io
//
// NOTE: This does come with around a 30% performance hit compared to the CGO version of the driver.
// however it greatly simplifies the build process and allows for cross compilation.
package cgofreesqlite

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/sysadminsmedia/homebox/backend/pkgs/textutils"
	"modernc.org/sqlite"
)

type CGOFreeSqliteDriver struct {
	*sqlite.Driver
}

type sqlite3DriverConn interface {
	Exec(string, []driver.Value) (driver.Result, error)
}

func (d CGOFreeSqliteDriver) Open(name string) (conn driver.Conn, err error) {
	conn, err = d.Driver.Open(name)
	if err != nil {
		return nil, err
	}
	_, err = conn.(sqlite3DriverConn).Exec("PRAGMA foreign_keys = ON;", nil)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, err
}

// modernDriver returns modernc's package-level driver singleton (the instance
// it registers as "sqlite"). Functions registered through the sqlite package
// (like hb_fold below) are stored on that singleton only, so wrapping a fresh
// &sqlite.Driver{} would silently lose them.
func modernDriver() *sqlite.Driver {
	db, err := sql.Open("sqlite", "")
	if err != nil {
		panic(err)
	}
	defer func() { _ = db.Close() }()
	return db.Driver().(*sqlite.Driver)
}

func init() { //nolint:gochecknoinits
	sql.Register("sqlite3", CGOFreeSqliteDriver{Driver: modernDriver()})

	// hb_fold(text) folds its argument for case- and accent-insensitive
	// comparison (full Unicode case folding + diacritic removal). SQLite's
	// built-in lower()/LIKE only handle ASCII, which breaks search for
	// Cyrillic, Greek, and other non-ASCII scripts. The search engine compares
	// hb_fold(column) against patterns folded the same way in Go.
	sqlite.MustRegisterDeterministicScalarFunction("hb_fold", 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		switch v := args[0].(type) {
		case nil:
			return nil, nil
		case string:
			return textutils.Fold(v), nil
		case []byte:
			return textutils.Fold(string(v)), nil
		default:
			return nil, fmt.Errorf("hb_fold: unsupported argument type %T", v)
		}
	})
}
