package radioslack

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/antonholmquist/jason"
	"github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"github.com/parnurzeal/gorequest"
)

type ErrorJson struct {
	Error string `json:"error"`
}

type ChannelJson struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type MeJson struct {
	TeamId   string        `json:"team_id"`
	Team     string        `json:"team"`
	User     string        `json:"user"`
	Channels []ChannelJson `json:"channels"`
}

func Me(c *echo.Context) error {
	cookie, err := c.Request().Cookie("profile")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorJson{
			Error: "Unauthorized",
		})
	}
	profile, _ := base64.StdEncoding.DecodeString(cookie.Value)
	v, _ := jason.NewObjectFromBytes(profile)
	teamId, _ := v.GetString("team_id")
	team, _ := v.GetString("team")
	user, _ := v.GetString("user")
	rc := rp.Get()
	token, _ := redis.String(
		rc.Do("GET", fmt.Sprintf("radioslack:%s:token", teamId)),
	)
	req := gorequest.New()
	res, _, _ := req.
		Get("https://slack.com/api/channels.list").
		Query("token=" + token).
		End()
	v, _ = jason.NewObjectFromReader(res.Body)
	channels := []ChannelJson{}
	_channels, _ := v.GetObjectArray("channels")
	for _, channel := range _channels {
		chId, _ := channel.GetString("id")
		chName, _ := channel.GetString("name")
		channels = append(channels, ChannelJson{
			Id:   chId,
			Name: chName,
		})
	}
	me := MeJson{
		TeamId:   teamId,
		Team:     team,
		User:     user,
		Channels: channels,
	}
	return c.JSON(http.StatusOK, me)
}
