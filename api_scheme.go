package surrealdbdriver

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/senpro-it/dsb-tool/extras/surrealdb-driver/surrealtypes"
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

type SurrealAPIError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type SurrealAPIResponse struct {
	ID     SurrealRequestID `json:"id"`
	Result any              `json:"result"`
	// Usually returned from HTTP errors
	Code        int    `json:"code,omitempty"`
	Details     string `json:"details,omitempty"`
	Information string `json:"information,omitempty"`
	// Actual errors
	Error SurrealAPIError `json:"error,omitempty"`
}

/*func (r *SurrealAPIResponse) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Check if the "error" field exists
	if _, hasError := raw["error"]; hasError {
		log.Println("unmarshalling found an error...")
		// Temporary re-type to avoid re-calling this method.
		type errorOnly SurrealAPIResponse
		var errResponse errorOnly
		if err := json.Unmarshal(data, &errResponse); err != nil {
			return err
		}
		*r = SurrealAPIResponse{
			ID:    errResponse.ID, // ID is always present
			Error: errResponse.Error,
		}
		return nil
	}

	// Same re-type.
	log.Println("unmarshalling found NO error!")
	type normalResponse SurrealAPIResponse
	var successResponse normalResponse
	if err := json.Unmarshal(data, &successResponse); err != nil {
		return err
	}
	*r = SurrealAPIResponse(successResponse)

	return nil
}*/

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
	ID     int         `json:"id,omitempty"`
	Result interface{} `json:"result"`
	Status string      `json:"status,omitempty"`
	Time   string      `json:"time,omitempty"`
}

type SurrealInfoQueryResponse struct {
	Accesses  surrealtypes.StringMap `json:"accesses"`
	Analyzers surrealtypes.StringMap `json:"analyzers"`
	Configs   surrealtypes.StringMap `json:"configs"`
	Functions surrealtypes.StringMap `json:"functions"`
	Models    surrealtypes.StringMap `json:"models"`
	Params    surrealtypes.StringMap `json:"params"`
	Tables    surrealtypes.StringMap `json:"tables"`
	Users     surrealtypes.StringMap `json:"users"`
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
func (c *SurrealCaller) CallUse(ns string, db string) *SurrealAPIRequest {
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
	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "use",
		Params: params,
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
func (c *SurrealCaller) CallSignin(creds *CredentialConfig) (*SurrealAPIRequest, error) {
	params := map[string]interface{}{}
	switch creds.Method {
	case AuthMethodRoot:
		params["user"] = creds.Username
		params["pass"] = creds.Password
	case AuthMethodRecord:
		params["user"] = creds.Username
		params["pass"] = creds.Password
		params["NS"] = creds.Namespace
		params["DB"] = creds.Database
		params["AC"] = creds.AccessControl
	case AuthMethodDB:
		params["user"] = creds.Username
		params["pass"] = creds.Password
		params["NS"] = creds.Namespace
		params["DB"] = creds.Database
	case AuthMethodUnknown:
		return nil, errors.New("unknown authentication method")
	case AuthMethodToken:
		// nop: Must use different API call (identify)
	case AuthMethodAnonymous:
		// nop: No credentials needed (i.e. for mem:// storage)
	default:
		return nil, errors.New("unrecognized authentication method: " + string(creds.Method))
	}

	if len(creds.extra) > 0 {
		for k, v := range creds.extra {
			params[k] = v
		}
	}

	return &SurrealAPIRequest{
		ID:     c.ConnID,
		Method: "signin",
		Params: []any{params},
	}, nil
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
