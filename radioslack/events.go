package radioslack

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
)

func Events(c *echo.Context) error {
	teamId := c.P(0)
	ch := c.P(1)
	ws := c.Socket()
	rc := rp.Get()
	psc := redis.PubSubConn{rc}
	psc.Subscribe(fmt.Sprintf("radioslack:%s:%s:songs", teamId, ch))
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			websocket.Message.Send(ws, string(v.Data))
		case error:
			return v
		}
	}
}
