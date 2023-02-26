package controller

import (
	"errors"
	"fmt"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/entity"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/service"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12/sessions"
	"time"
)

/**
使用iris自带的会话管理功能
*/
var expire, _ = config.GetInt("session.expire", 45)
var (
	cookieNameForSessionID = "irissessionid"
	sess                   = sessions.New(sessions.Config{Cookie: cookieNameForSessionID,
		Expires: time.Duration(expire) * time.Minute, // <=0 意味永久的存活
	})
)

func GetSession() *sessions.Sessions {
	return sess
}

func init() {
	db := NewDatabase()
	if db != nil {
		sess.UseDatabase(db)
	} else {
		panic("session new database fail")
	}

	//这是badger
	//db, err := badger.New("")
	//if err == nil {
	//	sess.UseDatabase(db)
	//}

	//这是redis
	//cfg:=redis.Config{}
	//db = redis.New(cfg)
	//if err != nil {
	//	sess.UseDatabase(db)
	//}
}

/**
会话数据的数据库集中存储XORM，适用于集群环境的集中会话管理，有另外的两个实现
分别是redis和badger，都是实现了sessions.Database接口，并设置session使用这个接口的实现
*/
type XORMDatabase struct {
	sessionService     *service.SessionService
	sessionDataService *service.SessionDataService
	logger             *golog.Logger
}

var _ sessions.Database = (*XORMDatabase)(nil)

func NewDatabase() sessions.Database {
	return &XORMDatabase{
		sessionService:     service.GetSessionService(),
		sessionDataService: service.GetSessionDataService(),
	}
}

// SetLogger sets the logger once before server ran.
// By default the Iris one is injected.
func (db *XORMDatabase) SetLogger(logger *golog.Logger) {
	db.logger = logger
}

/**
没有实现完毕，没有测试
*/
// Acquire receives a session's lifetime from the database,
// if the return value is LifeTime{} then the session manager sets the life time based on the expiration duration lives in configuration.
func (db *XORMDatabase) Acquire(sid string, expires time.Duration) sessions.LifeTime {
	session := entity.Session{SessionId: sid}
	found, _ := db.sessionService.Get(&session, false, "", "")
	if found {
		lifeTime := session.LifeTime
		return sessions.LifeTime{
			Time: time.Now().Add(time.Duration(lifeTime) * time.Second),
		}
	} else {
		db.sessionService.Insert(&session)

		return sessions.LifeTime{}
	}
}

// OnUpdateExpiration will re-set the database's session's entry ttl.
// https://redis.io/commands/expire#refreshing-expires
func (db *XORMDatabase) OnUpdateExpiration(sid string, newExpires time.Duration) error {
	session := entity.Session{SessionId: sid}
	found, _ := db.sessionService.Get(&session, false, "", "")
	if found {
		session.LifeTime = int64(newExpires.Seconds())
		mds := make([]interface{}, 1)
		mds[0] = session
		db.sessionService.Update(mds, nil, "")

		return nil
	} else {
		return errors.New("NotFound")
	}
}

// Set sets a key value of a specific session.
// Ignore the "immutable".
func (db *XORMDatabase) Set(sid string, key string, value interface{}, ttl time.Duration, immutable bool) error {
	val := ""
	if value == nil {
		val = fmt.Sprintf("%v", value)
	}
	sessionData := entity.SessionData{
		SessionId: sid,
		Key:       key,
	}
	found, _ := db.sessionDataService.Get(&sessionData, false, "", "")
	bs, err := sessions.DefaultTranscoder.Marshal(value)
	if err == nil && val != "" {
		val = string(bs)
	}
	if err == nil && found {
		sessionData.Value = val
		sessionData.LifeTime = int64(ttl)
		mds := make([]interface{}, 1)
		mds[0] = sessionData
		db.sessionDataService.Update(mds, nil, "")
	} else {
		sessionData.Value = val
		sessionData.LifeTime = int64(ttl)
		mds := make([]interface{}, 1)
		mds[0] = sessionData
		db.sessionDataService.Insert(mds, nil)
	}

	return nil
}

// Get retrieves a session value based on the key.
func (db *XORMDatabase) Get(sid string, key string) (value interface{}) {
	sessionData := entity.SessionData{
		SessionId: sid,
		Key:       key,
	}
	found, _ := db.sessionDataService.Get(&sessionData, false, "", "")
	if found {
		return sessionData.Value
	} else {
		return nil
	}
}

func (db *XORMDatabase) get(sid string, key string, outPtr interface{}) {
	data := db.Get(sid, key)
	if data != nil {
		err := sessions.DefaultTranscoder.Unmarshal(data.([]byte), outPtr)
		if err != nil {
			logger.Sugar.Errorf("%v", err)
		}
	}
	logger.Sugar.Errorf("%v", "NotFound")
}

func (db *XORMDatabase) keys(sid string) []string {
	rowsSlicePtr := make([]entity.SessionData, 0)
	sessionData := entity.SessionData{
		SessionId: sid,
	}
	db.sessionDataService.Find(&rowsSlicePtr, &sessionData, "", 0, 0, "")
	keys := make([]string, len(rowsSlicePtr))
	for i := 0; i < len(rowsSlicePtr); i++ {
		keys[i] = rowsSlicePtr[i].Key
	}
	return keys
}

// Visit loops through all session keys and values.
func (db *XORMDatabase) Visit(sid string, cb func(key string, value interface{})) error {
	keys := db.keys(sid)
	for _, key := range keys {
		var value interface{} // new value each time, we don't know what user will do in "cb".
		db.get(sid, key, &value)
		cb(key, value)
	}

	return nil
}

// Len returns the length of the session's entries (keys).
func (db *XORMDatabase) Len(sid string) (n int) {
	return len(db.keys(sid))
}

// Delete removes a session key value based on its key.
func (db *XORMDatabase) Delete(sid string, key string) (deleted bool) {
	sessionData := entity.SessionData{
		SessionId: sid,
		Key:       key,
	}
	mds := make([]interface{}, 1)
	mds[0] = sessionData
	affected, _ := db.sessionDataService.Delete(mds, "")
	if affected != 0 {
		return true
	} else {
		return false
	}
}

// Clear removes all session key values but it keeps the session entry.
func (db *XORMDatabase) Clear(sid string) error {
	sessionData := entity.SessionData{
		SessionId: sid,
	}
	mds := make([]interface{}, 1)
	mds[0] = sessionData
	affected, _ := db.sessionDataService.Delete(mds, "")
	if affected == 0 {
		logger.Sugar.Errorf("NoDeleted")
	}

	return nil
}

// Release destroys the session, it clears and removes the session entry,
// session manager will create a new session ID on the next request after this call.
func (db *XORMDatabase) Release(sid string) error {
	// clear all $sid-$key.
	db.Clear(sid)
	// and remove the $sid.
	session := entity.Session{
		SessionId: sid,
	}
	mds := make([]interface{}, 1)
	mds[0] = session
	affected, _ := db.sessionService.Delete(mds, "")
	if affected >= 0 {
		db.logger.Debugf("Database.Release.Driver.Delete: %s: %v", sid, affected)
	}

	return nil
}

// Close terminates the redis connection.
func (db *XORMDatabase) Close() error {
	return nil
}

func (db *XORMDatabase) Decode(sid string, key string, outPtr interface{}) error {
	return nil
}
