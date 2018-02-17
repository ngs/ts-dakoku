package app

import (
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
)

func (app *App) reconnectRedisIfNeeeded() {
	res, _ := app.RedisConn.Do("PING")
	if pong, ok := res.([]byte); !ok || string(pong) != "PONG" {
		err := app.setupRedis()
		if err != nil {
			panic(err)
		}
	}
}

func (app *App) setupRedis() error {
	connectTimeout := 1 * time.Second
	readTimeout := 1 * time.Second
	writeTimeout := 1 * time.Second

	if url := os.Getenv("REDIS_URL"); url != "" {
		conn, err := redis.DialURL(url,
			redis.DialConnectTimeout(connectTimeout),
			redis.DialReadTimeout(readTimeout),
			redis.DialWriteTimeout(writeTimeout))
		if err != nil {
			return err
		}
		app.RedisConn = conn
		return nil
	}
	conn, err := redis.Dial("tcp", ":6379",
		redis.DialConnectTimeout(connectTimeout),
		redis.DialReadTimeout(readTimeout),
		redis.DialWriteTimeout(writeTimeout))
	if err != nil {
		return err
	}
	app.RedisConn = conn
	return nil
}
