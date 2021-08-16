package taossqlrestful

import "database/sql/driver"

type taosSqlStmt struct {
	mc         *taosConn
	id         uint32
	pSql       string
	paramCount int
}

func (stmt *taosSqlStmt) Close() error {
	return nil
}

func (stmt *taosSqlStmt) NumInput() int {
	return stmt.paramCount
}

func (stmt *taosSqlStmt) Exec(args []driver.Value) (driver.Result, error) {
	if stmt.mc == nil || stmt.mc.taos == "" {
		return nil, errInvalidConn
	}
	return stmt.mc.Exec(stmt.pSql, args)
}

func (stmt *taosSqlStmt) Query(args []driver.Value) (driver.Rows, error) {
	if stmt.mc == nil || stmt.mc.taos == "" {
		return nil, errInvalidConn
	}
	return stmt.query(args)
}

func (stmt *taosSqlStmt) query(args []driver.Value) (*taosSqlRows, error) {
	mc := stmt.mc
	if mc == nil || mc.taos == "" {
		return nil, errInvalidConn
	}

	querySQL := stmt.pSql

	if len(args) != 0 {
		if !mc.cfg.interpolateParams {
			return nil, driver.ErrSkip
		}
		// try client-side prepare to reduce roundtrip
		prepared, err := mc.interpolateParams(stmt.pSql, args)
		if err != nil {
			return nil, err
		}
		querySQL = prepared
	}

	numFields, err := mc.taosQuery(querySQL)
	if err == nil {
		// Read Result
		rows := new(taosSqlRows)
		rows.mc = mc
		// Columns field
		rows.rs.columns, err = mc.readColumns(numFields)
		return rows, err
	}
	return nil, err
}

// type converter struct{}

// func (c converter) ConvertValue(v interface{}) (driver.Value, error) {

// 	if driver.IsValue(v) {
// 		return v, nil
// 	}

// 	if vr, ok := v.(driver.Valuer); ok {
// 		sv, err := callValuerValue(vr)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if !driver.IsValue(sv) {
// 			return nil, fmt.Errorf("non-Value type %T returned from Value", sv)
// 		}

// 		return sv, nil
// 	}

// 	rv := reflect.ValueOf(v)
// 	switch rv.Kind() {
// 	case reflect.Ptr:
// 		// indirect pointers
// 		if rv.IsNil() {
// 			return nil, nil
// 		} else {
// 			return c.ConvertValue(rv.Elem().Interface())
// 		}
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		return rv.Int(), nil
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 		return rv.Uint(), nil
// 	case reflect.Float32, reflect.Float64:
// 		return rv.Float(), nil
// 	case reflect.Bool:
// 		return rv.Bool(), nil
// 	case reflect.Slice:
// 		ek := rv.Type().Elem().Kind()
// 		if ek == reflect.Uint8 {
// 			return rv.Bytes(), nil
// 		}
// 		return nil, fmt.Errorf("unsupported type %T, a slice of %s", v, ek)
// 	case reflect.String:
// 		return rv.String(), nil
// 	}
// 	return nil, fmt.Errorf("unsupported type %T, a %s", v, rv.Kind())
// }

// var valuerReflectType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

// func callValuerValue(vr driver.Valuer) (v driver.Value, err error) {
// 	if rv := reflect.ValueOf(vr); rv.Kind() == reflect.Ptr &&
// 		rv.IsNil() &&
// 		rv.Type().Elem().Implements(valuerReflectType) {
// 		return nil, nil
// 	}
// 	return vr.Value()
// }
