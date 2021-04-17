package status

var redisDone = make(chan bool)

var mysqlDone = make(chan bool)

var telnetDone = make(chan bool)

var sshDone = make(chan bool)

var webDone = make(chan bool)

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

func GetSshDone() chan bool {
	return sshDone
}

func SetSshDone(done bool) {
	sshDone <- done
}

func GetWebDone() chan bool {
	return webDone
}

func SetWebDone(done bool) {
	webDone <- done
}

var redisStatus = false

var mysqlStatus = false

var telnetStatus = false

var sshStatus = false

var webStatus = false

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

func SetSshStatus(status bool) {
	sshStatus = status
}

func GetSshStatus() bool {
	return sshStatus
}

func SetWebStatus(status bool) {
	webStatus = status
}

func GetWebStatus() bool {
	return webStatus
}

var redisUnMatch = make(chan string)

func SetRedisUnMatch(unMatchId string) {
	redisUnMatch <- unMatchId
}

func GetRedisUnMatch() chan string {
	return redisUnMatch
}
