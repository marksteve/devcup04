package radioslack

import (
	"fmt"
	"log"
	"net/http"

	"github.com/antonholmquist/jason"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/parnurzeal/gorequest"
)

var allowedServices = map[string]bool{
	"YouTube":    true,
	"SoundCloud": true,
}

func TeamRTM(teamId string) {
	for {
		// TODO: Locking!
		rc := rp.Get()
		token, _ := redis.String(rc.Do(
			"GET",
			fmt.Sprintf("radioslack:%s:token", teamId),
		))
		req := gorequest.New()
		res, _, _ := req.
			Get("https://slack.com/api/rtm.start").
			Query("token=" + token).
			End()
		v, _ := jason.NewObjectFromReader(res.Body)
		u, _ := v.GetString("url")
		log.Printf("Starting RTM: %s", u)
		ws, _, _ := websocket.DefaultDialer.Dial(u, http.Header{})
		for {
			_, msg, _ := ws.ReadMessage()
			v, _ = jason.NewObjectFromBytes(msg)
			t, _ := v.GetString("type")
			if t != "message" {
				continue
			}
			ch, _ := v.GetString("channel")
			attachments, _ := v.GetObjectArray("message", "attachments")
			for _, attachment := range attachments {
				sn, _ := attachment.GetString("service_name")
				_, allowed := allowedServices[sn]
				if !allowed {
					continue
				}
				_rc := rp.Get()
				key := fmt.Sprintf("radioslack:%s:%s:songs", teamId, ch)
				_rc.Do("SET", key, attachment.String())
				_rc.Do("PUBLISH", key, "new_song")
				log.Printf("New message on: %s", key)
			}
		}
	}
}

func Start() error {
	for {
		rc := rp.Get()
		teams, _ := redis.Strings(rc.Do("SMEMBERS", "radioslack:teams"))
		for _, teamId := range teams {
			log.Printf("Existing team: %s", teamId)
			go TeamRTM(teamId)
		}
		psc := redis.PubSubConn{rc}
		ch := "radioslack"
		psc.Subscribe(ch)
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				teamId := string(v.Data)
				log.Printf("New team: %s", teamId)
				go TeamRTM(teamId)
			case error:
				return v
			}
		}
	}
}
