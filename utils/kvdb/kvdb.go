package kvdb

type KvDb interface {
	Get(key string) (val any, err error)
	GetInt(key string) (i int32, err error)
	GetInt64(key string) (i64 int64, err error)
	GetString(key string) (str string, err error)
	GetBool(key string) (b bool, err error)
	Set(key string, value any) (err error)
	SetInt(key string, value int) (err error)
	SetInt64(key string, value int64) (err error)
	SetString(key string, value string) (err error)
	SetBool(key string, value bool) (err error)
	Delete(key string) (err error)
	Expire(key string, expireTime int64) (err error)
	ExpireAt(key string, expireTimestamp int64) (err error)
	Persist(key string) (err error)
}

var instance KvDb = nil

func GetKvDb() *KvDb {
	if instance != nil {
		return &instance
	}
	mem := CreateMemoryKvDb()
	instance = &mem
	return &instance
}

func init() {

}
