package ent

import (
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

// Sql exposes the underlying database connection in the ent client
// so that we can use it to perform custom queries.
func (c *Client) Sql() *sql.DB {
	switch drv := c.driver.(type) {
	case *entsql.Driver:
		return drv.DB()
	case *dialect.DebugDriver:
		return drv.Driver.(*entsql.Driver).DB()
	default:
		panic(fmt.Errorf("unsupported driver type %T", drv))
	}
}
