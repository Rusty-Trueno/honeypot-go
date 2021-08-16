package taossqlrestful

import (
	"context"
	"database/sql/driver"
)

type connector struct {
	cfg *config
}

// Connect implements driver.Connector interface.
// Connect returns a connection to the database.
func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	var err error
	// New taosConn
	mc := &taosConn{
		cfg:       c.cfg,
		parseTime: c.cfg.parseTime,
	}

	// Connect to Server
	if len(mc.cfg.user) == 0 {
		mc.cfg.user = "root"
	}
	if len(mc.cfg.passwd) == 0 {
		mc.cfg.passwd = "taosdata"
	}
	mc.taos, err = mc.taosConnect(mc.cfg.addr, mc.cfg.user, mc.cfg.passwd, mc.cfg.dbName, mc.cfg.port)
	if err != nil {
		return nil, err
	}
	// if len(mc.cfg.dbName) != 0 {
	// 	_, err := mc.Exec(strings.Join([]string{"use", mc.cfg.dbName}, " "), nil)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	//return nil, err
	// }
	return mc, nil
}

// Driver implements driver.Connector interface.
// Driver returns &taosSQLDriver{}.
func (c *connector) Driver() driver.Driver {
	return &restfulDriver{}
}
