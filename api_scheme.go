package surrealdbdriver

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type SurrealRequestID = string

func GenerateSecureID() SurrealRequestID {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

type SurrealAPIRequest struct {
	ID     SurrealRequestID `json:"id"`
	Method string           `json:"method"`
	Params interface{}      `json:"params,omitempty"`
}

type SurrealAPIResponse struct {
	ID     SurrealRequestID `json:"id"`
	Result any              `json:"result"`
	// Usually returned from HTTP errors
	Code        int    `json:"code,omitempty"`
	Details     string `json:"details,omitempty"`
	Information string `json:"information,omitempty"`
	// Actual errors
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// A few predefined structs:
type SurrealVersion struct {
	Version   string    `json:"version"`
	Build     string    `json:"build"`
	Timestamp time.Time `json:"timestamp"`
}
type SurrealLiveNotification struct {
	Action string      `json:"action"`
	ID     string      `json:"id"`
	Result interface{} `json:"result"`
}
type SurrealQueryResponse struct {
	ID     int         `json:"id"`
	Result interface{} `json:"result"`
	Status string      `json:"status,omitempty"`
	Time   string      `json:"time,omitempty"`
}

type SurrealCaller struct {
	ConnID SurrealRequestID
}

func MakeCaller() *SurrealCaller {
	return &SurrealCaller{
		ConnID: GenerateSecureID(),
	}
}

func (c *SurrealCaller) CallVersion() *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "version",
		Params: nil,
	}
}
func (c *SurrealCaller) CallUse(db string, ns string) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "use",
		Params: []string{ns, db},
	}
}
func (c *SurrealCaller) CallInfo() *SurrealAPIRequest {
	return &SurrealAPIRequest{
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
) *SurrealAPIRequest {
	params := map[string]string{
		"NS": ns,
		"DB": db,
		"AC": ac,
	}
	for k, v := range extra {
		params[k] = v
	}
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "signup",
		Params: []any{params},
	}
}
func (c *SurrealCaller) CallSignin(
	ns string,
	db string,
	ac string,
	username string,
	password string,
	extra map[string]string,
) *SurrealAPIRequest {
	params := map[string]string{
		"NS":   ns,
		"DB":   db,
		"AC":   ac,
		"user": username,
		"pass": password,
	}
	for k, v := range extra {
		params[k] = v
	}
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "signin",
		Params: []any{params},
	}
}
func (c *SurrealCaller) CallAuthenticate(token string) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "authenticate",
		Params: []string{token},
	}
}
func (c *SurrealCaller) CallInvalidate() *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "invalidate",
	}
}
func (c *SurrealCaller) CallLet(key string, value any) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "let",
		Params: []any{key, value},
	}
}
func (c *SurrealCaller) CallUnset(key string) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "unset",
		Params: []string{key},
	}
}
func (c *SurrealCaller) CallLive(table string, diff bool) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "live",
		Params: []any{table, diff},
	}
}
func (c *SurrealCaller) CallKill(queryUuid string) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "kill",
		Params: []string{queryUuid},
	}
}
func (c *SurrealCaller) CallQuery(
	sql string,
	vars map[string]interface{},
) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "query",
		Params: []any{sql, vars},
	}
}
func (c *SurrealCaller) CallRun(
	func_name string,
	version string,
	args []interface{},
) *SurrealAPIRequest {
	var params []interface{}
	params = append(params, version, args)
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "run",
		Params: params,
	}
}
func (c *SurrealCaller) CallGraphQL(
	query string,
	options map[string]interface{},
) *SurrealAPIRequest {
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
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "graphql",
		Params: params,
	}
}
func (c *SurrealCaller) CallSelect(thing string) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "select",
		Params: []string{thing},
	}
}
func (c *SurrealCaller) CallCreate(thing string, data interface{}) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "create",
		Params: []any{thing, data},
	}
}
func (c *SurrealCaller) CallInsert(thing string, data any) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "insert",
		Params: []any{thing, data},
	}
}
func (c *SurrealCaller) CallInsertRelation(
	thing string,
	data map[string]interface{},
) *SurrealAPIRequest {
	var params []interface{}
	if _, ok := data["id"]; ok && thing == "" {
		params = append(params, data)
	} else {
		params = append(params, thing)
		params = append(params, data)
	}
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "insert",
		Params: params,
	}
}
func (c *SurrealCaller) CallUpdate(thing string, data interface{}) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "update",
		Params: []any{thing, data},
	}
}
func (c *SurrealCaller) CallUpsert(thing string, data interface{}) *SurrealAPIRequest {
	return &SurrealAPIRequest{
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
) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "relate",
		Params: []any{in, relation, out, data},
	}
}
func (c *SurrealCaller) CallMerge(thing string, data interface{}) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "merge",
		Params: []any{thing, data},
	}
}
func (c *SurrealCaller) CallPatch(thing string, patches []interface{}, diff bool) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "merge",
		Params: []any{thing, patches, diff},
	}
}
func (c *SurrealCaller) CallDelete(thing string) *SurrealAPIRequest {
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "delete",
		Params: []string{thing},
	}
}
