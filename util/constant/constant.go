package constant

const (
	//honeypot type
	Redis  = "redis"
	Mysql  = "mysql"
	Ssh    = "ssh"
	Telnet = "telnet"
	Web    = "web"

	//kubeedge topic
	NodeETPrefix               = "$hw/events/node/"
	DeviceETMemberGetSuffix    = "/membership/get"
	DeviceETMemberResultSuffix = "/membership/get/result"
	DeviceETMemberUpdated      = "/membership/updated"
	DeviceETPrefix             = "$hw/events/device/"
	DeviceETStateUpdateSuffix  = "/state/update"
	TwinETUpdateSuffix         = "/twin/update"
	TwinETCloudSyncSuffix      = "/twin/cloud_updated"
	TwinETGetResultSuffix      = "/twin/get/result"
	TwinETGetSuffix            = "/twin/get"
	TwinETDeltaSuffix          = "/twin/update/delta"
	Switch                     = "switch"
	ModelName                  = "honeypot"

	//bypass topic
	BypassPotMsg = "/honeypot/msg/"

	//environment
	Windows = "windows"
	Linux   = "linux"
)
