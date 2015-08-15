package radioslack

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/antonholmquist/jason"
	"github.com/labstack/echo"
	"github.com/parnurzeal/gorequest"
)

func Login(c *echo.Context) error {
	u, _ := url.Parse("https://slack.com/oauth/authorize")
	q := u.Query()
	q.Set("client_id", c.Get("SlackClientId").(string))
	q.Set("scope", "client")
	q.Set("state", "ofthenation")
	u.RawQuery = q.Encode()
	return c.Redirect(http.StatusTemporaryRedirect, u.String())
}

func OAuth(c *echo.Context) error {
	code := c.Query("code")
	req := gorequest.New()
	res, _, _ := req.
		Get("https://slack.com/api/oauth.access").
		Query("client_id=" + c.Get("SlackClientId").(string)).
		Query("client_secret=" + c.Get("SlackClientSecret").(string)).
		Query("code=" + code).
		End()
	v, _ := jason.NewObjectFromReader(res.Body)
	token, _ := v.GetString("access_token")
	res, _, _ = req.
		Get("https://slack.com/api/auth.test").
		Query("token=" + token).
		End()
	v, _ = jason.NewObjectFromReader(res.Body)
	teamId, _ := v.GetString("team_id")
	rc := rp.Get()
	rc.Do("SADD", "radioslack:teams", teamId)
	rc.Do("SET", fmt.Sprintf("radioslack:%s:token", teamId), token)
	rc.Do("PUBLISH", "radioslack", teamId)
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
