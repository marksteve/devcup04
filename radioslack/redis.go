package radioslack

import (
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

var rp *redis.Pool

func init() {
	rp = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:6379", config.RedisHost))
			if err != nil {
				log.Fatal(err.Error())
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
