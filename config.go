package surrealdbdriver

import (
	"errors"
	"net/url"
)

type CredentialConfig struct {
	Username  string
	Password  string
	Database  string
	Namespace string
	Token     string
	URL       *url.URL
}

func (c *CredentialConfig) GetDBUrl() string {
	plainUrl := url.URL{
		Scheme:  c.URL.Scheme,
		Host:    c.URL.Host,
		RawPath: c.URL.RawPath,
	}
	return plainUrl.String()
}

func ParseUrl(inputUrl string) (*CredentialConfig, error) {
	u, err := url.Parse(inputUrl)
	if err != nil {
		return nil, err
	}
	c := &CredentialConfig{}
	q := u.Query()
	if q.Has("token") {
		c.Token = q.Get("token")
	}
	if q.Has("ns") {
		c.Namespace = q.Get("ns")
	}
	if q.Has("db") {
		c.Database = q.Get("db")
	}

	if !q.Has("token") {
		user := u.User
		username := user.Username()
		password, isSet := user.Password()
		if username != "" && isSet {
			c.Username = username
			c.Password = password
		} else {
			return nil, errors.New("No token and no username/password set")
		}
	}

	c.URL = u
	return c, nil
}
