package status

var redisDone = make(chan bool)

var mysqlDone = make(chan bool)

var telnetDone = make(chan bool)

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

func GetTelnetDone() chan bool {
	return telnetDone
}

func SetTelnetDone(done bool) {
	telnetDone <- done
}

var redisStatus = false

var mysqlStatus = false

var telnetStatus = false

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

func SetTelnetStatus(status bool) {
	telnetStatus = status
}

func GetTelnetStatus() bool {
	return telnetStatus
}

var redisUnMatch = make(chan string)

func SetRedisUnMatch(unMatchId string) {
	redisUnMatch <- unMatchId
}

func GetRedisUnMatch() chan string {
	return redisUnMatch
}
