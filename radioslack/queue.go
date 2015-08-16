package radioslack

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
)

func Queue(c *echo.Context) error {
	teamId := c.P(0)
	ch := c.P(1)
	ws := c.Socket()
	key := fmt.Sprintf("radioslack:%s:%s:queue", teamId, ch)
	rc := rp.Get()
	queue, _ := redis.Strings(rc.Do("ZRANGEBYSCORE", key, "-inf", "inf"))
	for _, songKey := range queue {
		song := getSong(rc, songKey)
		websocket.Message.Send(ws, song)
	}
	psc := redis.PubSubConn{rc}
	psc.Subscribe(key)
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			websocket.Message.Send(ws, string(v.Data))
		case error:
			return v
		}
	}
}
