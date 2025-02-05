package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

type Credentials struct {
	Username      string
	Password      string
	Database      string
	Namespace     string
	AccessControl string
	Token         string
	Method        AuthMethod
	URL           *url.URL
	Extra         map[string]interface{}
}

func (c *Credentials) GetDBUrl() string {
	return fmt.Sprintf(
		"%s://%s%s",
		c.URL.Scheme,
		c.URL.Host,
		c.URL.EscapedPath(),
	)
}

func ParseUrl(inputUrl string) (*Credentials, error) {
	u, err := url.Parse(inputUrl)
	if err != nil {
		return nil, err
	}
	c := &Credentials{}
	q := u.Query()
	if !q.Has("method") {
		return nil, errors.New("no authentication method specified")
	}

	// Sanity check
	c.Method = AuthMethod(q.Get("method"))
	switch c.Method {
	case AuthMethodRoot:
		if _, ok := u.User.Password(); !ok && u.User.Username() != "" {
			return nil, errors.New("root authentication method specified but no username or password provided")
		}
	case AuthMethodDB:
		if _, ok := u.User.Password(); !ok && u.User.Username() != "" && !q.Has("db") && !q.Has("ns") {
			return nil, errors.New("database authentication method specified but no username, password, namespace and database provided")
		}
	case AuthMethodRecord:
		if _, ok := u.User.Password(); !ok && u.User.Username() != "" && !q.Has("db") && !q.Has("ns") && !q.Has("ac") {
			return nil, errors.New("record authentication method specified but no username, password, namespace, database and access method provided")
		}
	case AuthMethodToken:
		if !q.Has("token") {
			return nil, errors.New("token authentication method specified but no access token provided")
		}
	default:
		return nil, errors.New("unknown access method: " + string(c.Method))
	}

	// Assign
	c.Username = u.User.Username()
	c.Password, _ = u.User.Password()
	c.Database = q.Get("db")
	c.Namespace = q.Get("ns")
	c.AccessControl = q.Get("ac")
	c.Token = q.Get("token")

	if q.Has("extra") {
		extraStr := q.Get("extra")
		err := json.Unmarshal([]byte(extraStr), &(c.Extra))
		if err != nil {
			return nil, err
		}
	}

	c.URL = u
	return c, nil
}
