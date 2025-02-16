package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/IngwiePhoenix/surrealdb-driver/config"
	"github.com/wI2L/jsondiff"
)

type RequestID = string

func GenerateSecureID() RequestID {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

type Request struct {
	ID     RequestID   `json:"id"`
	Method APIMethod   `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

type SurrealCaller struct {
	ConnID RequestID
}

func MakeCaller() *SurrealCaller {
	return &SurrealCaller{
		ConnID: GenerateSecureID(),
	}
}

func (c *SurrealCaller) CallVersion() *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "version",
		Params: nil,
	}
}
func (c *SurrealCaller) CallUse(ns string, db string) *Request {
	// TODO: There is currently no "none" support...which is invalid JSON, too. O.o
	params := []interface{}{}
	if ns == "" {
		params = append(params, nil)
	} else {
		params = append(params, ns)
	}
	if db == "" {
		params = append(params, nil)
	} else {
		params = append(params, db)
	}
	return &Request{
		ID:     c.ConnID,
		Method: "use",
		Params: params,
	}
}
func (c *SurrealCaller) CallInfo() *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "info",
		Params: nil,
	}
}
func (c *SurrealCaller) CallSignup(
	ns string,
	db string,
	ac string,
	extra map[string]string,
) *Request {
	params := map[string]string{
		"NS": ns,
		"DB": db,
		"AC": ac,
	}
	for k, v := range extra {
		params[k] = v
	}
	return &Request{
		ID:     c.ConnID,
		Method: "signup",
		Params: []any{params},
	}
}
func (c *SurrealCaller) CallSignin(creds *config.Credentials) (*Request, error) {
	params := map[string]interface{}{}
	switch creds.Method {
	case config.AuthMethodRoot:
		params["user"] = creds.Username
		params["pass"] = creds.Password
	case config.AuthMethodRecord:
		params["user"] = creds.Username
		params["pass"] = creds.Password
		params["NS"] = creds.Namespace
		params["DB"] = creds.Database
		params["AC"] = creds.AccessControl
	case config.AuthMethodDB:
		params["user"] = creds.Username
		params["pass"] = creds.Password
		params["NS"] = creds.Namespace
		params["DB"] = creds.Database
	case config.AuthMethodUnknown:
		return nil, errors.New("unknown authentication method")
	case config.AuthMethodToken:
		// nop: Must use different API call (identify)
	case config.AuthMethodAnonymous:
		// nop: No credentials needed (i.e. for mem:// storage)
	default:
		return nil, errors.New("unrecognized authentication method: " + string(creds.Method))
	}

	if len(creds.Extra) > 0 {
		for k, v := range creds.Extra {
			params[k] = v
		}
	}

	return &Request{
		ID:     c.ConnID,
		Method: "signin",
		Params: []any{params},
	}, nil
}
func (c *SurrealCaller) CallAuthenticate(token string) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "authenticate",
		Params: []string{token},
	}
}
func (c *SurrealCaller) CallInvalidate() *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "invalidate",
	}
}
func (c *SurrealCaller) CallLet(key string, value any) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "let",
		Params: []any{key, value},
	}
}
func (c *SurrealCaller) CallUnset(key string) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "unset",
		Params: []string{key},
	}
}
func (c *SurrealCaller) CallLive(table string, diff bool) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "live",
		Params: []any{table, diff},
	}
}
func (c *SurrealCaller) CallKill(queryUuid string) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "kill",
		Params: []string{queryUuid},
	}
}
func (c *SurrealCaller) CallQuery(
	sql string,
	vars map[string]interface{},
) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "query",
		Params: []any{sql, vars},
	}
}
func (c *SurrealCaller) CallRun(
	func_name string,
	version string,
	args []interface{},
) *Request {
	var params []interface{}
	params = append(params, version, args)
	return &Request{
		ID:     c.ConnID,
		Method: "run",
		Params: params,
	}
}
func (c *SurrealCaller) CallGraphQL(
	query string,
	options map[string]interface{},
) *Request {
	var params []interface{}
	if len(options) > 0 {
		// build a complex version
		params = append(params, map[string]string{
			"query": query,
		})
		params = append(params, options)
	} else {
		// simple one
		params = append(params, query)
	}
	return &Request{
		ID:     c.ConnID,
		Method: "graphql",
		Params: params,
	}
}
func (c *SurrealCaller) CallSelect(thing string) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "select",
		Params: []string{thing},
	}
}
func (c *SurrealCaller) CallCreate(thing string, data interface{}) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "create",
		Params: []any{thing, data},
	}
}
func (c *SurrealCaller) CallInsert(thing string, data any) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "insert",
		Params: []any{thing, data},
	}
}
func (c *SurrealCaller) CallInsertRelation(
	thing string,
	data map[string]interface{},
) *Request {
	var params []interface{}
	if _, ok := data["id"]; ok && thing == "" {
		params = append(params, data)
	} else {
		params = append(params, thing)
		params = append(params, data)
	}
	return &Request{
		ID:     c.ConnID,
		Method: "insert",
		Params: params,
	}
}
func (c *SurrealCaller) CallUpdate(thing string, data interface{}) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "update",
		Params: []any{thing, data},
	}
}
func (c *SurrealCaller) CallUpsert(thing string, data interface{}) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "upsert",
		Params: []any{thing, data},
	}
}
func (c *SurrealCaller) CallRelate(
	in string,
	relation string,
	out string,
	data interface{},
) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "relate",
		Params: []any{in, relation, out, data},
	}
}
func (c *SurrealCaller) CallMerge(thing string, data interface{}) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "merge",
		Params: []any{thing, data},
	}
}
func (c *SurrealCaller) CallPatch(thing string, patches jsondiff.Patch, diff bool) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "merge",
		Params: []any{thing, patches, diff},
	}
}
func (c *SurrealCaller) CallDelete(thing string) *Request {
	return &Request{
		ID:     c.ConnID,
		Method: "delete",
		Params: []string{thing},
	}
}
