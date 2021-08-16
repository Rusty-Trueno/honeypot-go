package tdengine

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "honeypot/util/taosSqlRestful"
)

var (
	TdEngineDB *sql.DB
)

func init() {
	var err error
	TdEngineDB, err = sql.Open("taossqlrestful", "root:taosdata@/http(iot-cloud-db-tdengine:6041)/test")
	if err != nil {
		fmt.Errorf("open restful timeseries failed, error is %v\n", err)
	}
}

func InsertPotData(potName, potData string) error {
	stms, err := TdEngineDB.Prepare(fmt.Sprintf("insert into %s(ts, value) values(?,?)", potName))
	if err != nil {
		fmt.Printf("prepare insert sql failed, error is %v\n", err)
		return err
	}

	rs, err := stms.Exec(time.Now(), potData)
	if err != nil {
		fmt.Printf("insert data failed, error is %v\n", err)
	}

	//我们可以获得插入的id
	id, err := rs.LastInsertId()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(id)
	//可以获得影响行数
	affect, err := rs.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(affect)

	return nil
}

func GetPotData() error {
	rows, err := TdEngineDB.Query("select * from pot_redis")
	if err != nil {
		fmt.Errorf("query pot data failed, error is %v", err)
	}

	for rows.Next() {
		var times string
		var value string
		if err := rows.Scan(&times, &value); err != nil {
			fmt.Errorf("scan field failed, error is %v", err)
		} else {
			var _value map[string]interface{}
			err := json.Unmarshal([]byte(value), &_value)
			if err != nil {
				fmt.Errorf("unmarshal map failed, error is %v", err)
			}
			fmt.Printf("value is %v\n", _value)
		}
	}
	return nil
}
