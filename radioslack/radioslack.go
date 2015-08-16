package radioslack

import (
	"encoding/json"
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

func getSong(rc redis.Conn, songKey string) string {
	song, _ := redis.StringMap(rc.Do("HGETALL", songKey))
	songJson, _ := json.Marshal(song)
	return string(songJson[:])
}

func processAttachments(key, user string, attachments []*jason.Object) {
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
		songKey := fmt.Sprintf("radioslack:songs:%x", fu)
		song := []interface{}{songKey}
		for af, av := range attachment.Map() {
			avs, err := av.String()
			if err != nil {
				avn, _ := av.Number()
				avs = avn.String()
			}
			song = append(song, af, avs)
		}
		song = append(song, "user", user)
		rc.Do("HMSET", song...)
		rc.Do("ZADD", key, nextEnd, songKey)
		rc.Do("PUBLISH", key, getSong(rc, songKey))
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
			user, _ := v.GetString("message", "user")
			attachments, _ := v.GetObjectArray("message", "attachments")
			if len(attachments) > 0 {
				key := fmt.Sprintf("radioslack:%s:%s:queue", teamId, ch)
				processAttachments(key, user, attachments)
			}
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
