package mybase

type UsrInRedis struct {
	ProtoData      string `json:"proto_data"`       //base64加密过的数据
	ProtoDataOrgin []byte `json:"proto_data_orgin"` //proto 原始数据
	VerInRedis     int64  `json:"ver_in_redis"`     //redis中的版本，0表示刚从数据库写入，每次改写redis中的数据的时候都必须判断版本是否大于这个值
	FlagSyncDB     int32  `json:"flag_sync_db"`     //是否已经将数据同步到数据库 0标识已经同步过，1标识尚未同步
	//WritingInRedis int32  `json:"writing_in_redis"` //正在写入redis的标志位，大于0表示有人正在写入。。0表示暂无人更新。
}

type UsrCheckIn struct {
	Day int8 `json:"day"` //签到第几天。
}
