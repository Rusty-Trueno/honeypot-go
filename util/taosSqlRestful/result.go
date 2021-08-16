package taossqlrestful

import (
	"database/sql/driver"
	"errors"
	"io"
	"time"
)

/******************************************************************************
*                              Result                                         *
******************************************************************************/
// Read Packets as Field Packets until EOF-Packet or an Error appears
func (mc *taosConn) readColumns(count int) ([]taosSqlField, error) {

	columns := make([]taosSqlField, count)

	if mc.result == nil {
		return nil, errors.New("invalid result")
	}

	for i := 0; i < count; i++ {
		columns[i].name = mc.result.Head[i]
		// columns[i].length = (uint32)(fields[i].bytes)
		// columns[i].fieldType = fieldType(fields[i]._type)
		// columns[i].flags = 0
		// columns[i].decimals  = 0
		//columns[i].charSet    = 0
	}
	return columns, nil
}

func (rows *taosSqlRows) readRow(dest []driver.Value) error {
	mc := rows.mc

	if rows.rs.done || mc == nil {
		return io.EOF
	}

	if mc.result == nil {
		return errors.New("result is nil! ")
	}

	if rows.rs.index >= int64(mc.result.Rows) {
		rows.rs.done = true
		mc.result = nil
		rows.mc = nil
		return io.EOF
	}
	//TODO: 字段中有其他类型的字段需要获取字段类型进行匹配
	for i := range dest {
		inter := mc.result.Data[rows.rs.index][i]
		// TODO:取第一个字段转换有问题
		// if i == 0 {
		// 	if mc.cfg.parseTime == true {
		// 		timestamp := int64(inter.(float64))
		// 		dest[i] = timestampConvertToString(timestamp, 0)
		// 	} else {
		// 		dest[i] = int64(inter.(float64))
		// 	}
		// } else {
		switch inter.(type) {
		case string:
			dest[i] = inter.(string)
			break
		case int:
			dest[i] = inter.(int64)
			break
		case float64:
			dest[i] = inter.(float64)
			break
		default:
			// fmt.Println("default fieldType: set dest[] to nil")
			dest[i] = nil
			break
		}
		// }

	}
	rows.rs.index++
	return nil
}

func timestampConvertToString(timestamp int64, precision int) string {
	switch precision {
	case 0: // milli-second
		s := timestamp / 1e3
		ns := timestamp % 1e3 * 1e6
		return time.Unix(s, ns).Format("2006-01-02 15:04:05.000")
	case 1: // micro-second
		s := timestamp / 1e6
		ns := timestamp % 1e6 * 1e3
		return time.Unix(s, ns).Format("2006-01-02 15:04:05.000000")
	case 2: // nano-second
		s := timestamp / 1e9
		ns := timestamp % 1e9
		return time.Unix(s, ns).Format("2006-01-02 15:04:05.000000000")
	default:
		panic("unknown precision")
	}
}
