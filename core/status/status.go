package status

var redisDone = make(chan bool)

var mysqlDone = make(chan bool)

func GetRedisDone() chan bool {
	return redisDone
}

func SetRedisDone(done bool) {
	redisDone <- done
}

func GetMysqlDone() chan bool {
	return mysqlDone
}

func SetMysqlDone(done bool) {
	mysqlDone <- done
}

var redisStatus = false

var mysqlStatus = false

func SetRedisStatus(status bool) {
	redisStatus = status
}

func GetRedisStatus() bool {
	return redisStatus
}

func SetMysqlStatus(status bool) {
	mysqlStatus = status
}

func GetMysqlStatus() bool {
	return mysqlStatus
}
