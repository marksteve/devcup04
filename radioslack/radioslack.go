package radioslack

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/parnurzeal/gorequest"
)

var allowedServices = map[string]bool{
	"SoundCloud": true,
}

func processAttachments(key string, attachments []*jason.Object) {
	for _, attachment := range attachments {
		sn, _ := attachment.GetString("service_name")
		_, allowed := allowedServices[sn]
		if !allowed {
			continue
		}
		fu, _ := attachment.GetString("from_url")
		req := gorequest.New()
		res, _, _ := req.
			Get("http://api.soundcloud.com/resolve.json").
			Query("url=" + fu).
			Query("client_id=392fa845f8cff3705b90006915b15af0").
			End()
		v, _ := jason.NewObjectFromReader(res.Body)
		rc := rp.Get()
		now := time.Now().Unix()
		end, _ := redis.Int64(rc.Do("GET", key+":end"))
		du, _ := v.GetInt64("duration")
		du = du / 1000
		var nextEnd int64
		if end < now {
			nextEnd = now + du
		} else {
			nextEnd = end + du
		}
		rc.Do("ZADD", key, nextEnd, attachment.String())
		rc.Do("PUBLISH", key, attachment.String())
		rc.Do("SET", key+":end", nextEnd)
		rc.Do("SADD", "radioslack:queues", key)
	}
}

func TeamWorker(teamId string) {
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
			key := fmt.Sprintf("radioslack:%s:%s:queue", teamId, ch)
			processAttachments(key, attachments)
		}
	}
}

func QueueWorker() {
	for {
		rc := rp.Get()
		queues, _ := redis.Strings(rc.Do("SMEMBERS", "radioslack:queues"))
		for _, queueKey := range queues {
			rc.Do("ZREMRANGEBYSCORE", queueKey, 0, time.Now().Unix())
		}
		time.Sleep(5 * time.Second)
	}
}

func Start() error {
	for {
		rc := rp.Get()
		teams, _ := redis.Strings(rc.Do("SMEMBERS", "radioslack:teams"))
		for _, teamId := range teams {
			log.Printf("Existing team: %s", teamId)
			go TeamWorker(teamId)
		}

		go QueueWorker()

		psc := redis.PubSubConn{rc}
		ch := "radioslack"
		psc.Subscribe(ch)
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				teamId := string(v.Data)
				log.Printf("New team: %s", teamId)
				go TeamWorker(teamId)
			case error:
				return v
			}
		}
	}
}
