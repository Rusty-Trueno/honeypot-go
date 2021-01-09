package status

var redisStop = false

func GetRedisStatus() bool {
	return redisStop
}

func SetRedisStatus(status bool) {
	redisStop = status
}
