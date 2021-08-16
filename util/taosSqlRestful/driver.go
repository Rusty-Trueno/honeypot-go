package taossqlrestful

import (
	"context"
	"database/sql"
	"database/sql/driver"
)

// Driver mydb driver for implement database/sql/driver
type restfulDriver struct {
}

func init() {
	// log.Println("register taossqlrestful driver")
	sql.Register("taossqlrestful", &restfulDriver{})
}

// Open for implement driver interface
func (driver *restfulDriver) Open(name string) (driver.Conn, error) {
	cfg, err := parseDSN(name)
	if err != nil {
		return nil, err
	}
	c := &connector{
		cfg: cfg,
	}
	return c.Connect(context.Background())
}
