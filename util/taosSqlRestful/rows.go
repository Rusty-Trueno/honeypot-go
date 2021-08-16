package taossqlrestful

import (
	"database/sql/driver"
	"io"
)

// RowS  RowS implemmet for driver.Rows
// type RowS struct {
// 	Size int64
// }

type taosSqlField struct {
	tableName string
	name      string
	length    uint32
	flags     fieldFlag // indicate whether this field can is null
	fieldType fieldType
	decimals  byte
	charSet   uint8
}

type resultSet struct {
	columns     []taosSqlField
	columnNames []string
	done        bool
	index       int64
}

type taosSqlRows struct {
	mc *taosConn
	rs resultSet
}

// type binaryRows struct {
// 	taosSqlRows
// }

// type textRows struct {
// 	taosSqlRows
// }

func (rows *taosSqlRows) Columns() []string {
	if rows.rs.columnNames != nil {
		return rows.rs.columnNames
	}

	columns := make([]string, len(rows.rs.columns))
	if rows.mc != nil && rows.mc.cfg.columnsWithAlias {
		for i := range columns {
			if tableName := rows.rs.columns[i].tableName; len(tableName) > 0 {
				columns[i] = tableName + "." + rows.rs.columns[i].name
			} else {
				columns[i] = rows.rs.columns[i].name
			}
		}
	} else {
		for i := range columns {
			columns[i] = rows.rs.columns[i].name
		}
	}

	rows.rs.columnNames = columns

	return columns
}

// Close closes the rows iterator.
func (rows *taosSqlRows) Close() error {
	if rows.mc != nil {
		if rows.mc.result != nil {
			rows.mc.result = nil
		}
		rows.mc = nil
	}
	return nil
}

// Next is
func (rows *taosSqlRows) Next(dest []driver.Value) error {
	if mc := rows.mc; mc != nil {
		// Fetch next row from stream
		return rows.readRow(dest)
	}
	return io.EOF
}

func (rows *taosSqlRows) HasNextResultSet() (b bool) {
	if rows.mc == nil {
		return false
	}
	return rows.mc.status&statusMoreResultsExists != 0
}

func (rows *taosSqlRows) nextResultSet() (int, error) {
	if rows.mc == nil {
		return 0, io.EOF
	}

	// Remove unread packets from stream
	if !rows.rs.done {
		rows.rs.done = true
	}

	if !rows.HasNextResultSet() {
		rows.mc = nil
		return 0, io.EOF
	}
	rows.rs = resultSet{}
	return 0, nil
}

func (rows *taosSqlRows) nextNotEmptyResultSet() (int, error) {
	for {
		resLen, err := rows.nextResultSet()
		if err != nil {
			return 0, err
		}

		if resLen > 0 {
			return resLen, nil
		}

		rows.rs.done = true
	}
}

func (rows *taosSqlRows) NextResultSet() error {
	resLen, err := rows.nextNotEmptyResultSet()
	if err != nil {
		return err
	}

	rows.rs.columns, err = rows.mc.readColumns(resLen)
	return err
}
