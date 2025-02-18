package surrealdbdriver

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/tidwall/gjson"

	"github.com/IngwiePhoenix/surrealdb-driver/api"
	"github.com/IngwiePhoenix/surrealdb-driver/config"
	"github.com/clok/kemba"
	"github.com/gorilla/websocket"
)

// implements driver.Conn
type SurrealConn struct {
	WSClient *websocket.Conn
	Driver   *SurrealDriver
	Caller   *api.SurrealCaller
	creds    *config.Credentials
	k        *kemba.Kemba
	e        *Debugger
}

var _ driver.Conn = (*SurrealConn)(nil)
var _ driver.ConnBeginTx = (*SurrealConn)(nil)
var _ driver.ConnPrepareContext = (*SurrealConn)(nil)
var _ driver.NamedValueChecker = (*SurrealConn)(nil)
var _ driver.ValueConverter = (*SurrealConn)(nil)
var _ driver.Pinger = (*SurrealConn)(nil)
var _ driver.ExecerContext = (*SurrealConn)(nil)
var _ driver.QueryerContext = (*SurrealConn)(nil)

// Execute directly on the underlying WebSockets connection by utilizing the
// raw API objects.
func (con *SurrealConn) execObj(req *api.Request) (*api.Response, error) {
	k := con.k.Extend("execObj")
	k.Log("received", req)

	// For debugging only
	if err := con.WSClient.WriteJSON(req); con.e.Debug(err) {
		return nil, err
	}

	mtyp, msg, err := con.WSClient.ReadMessage()
	if con.e.Debug(err) {
		return nil, err
	} else if mtyp != websocket.BinaryMessage && mtyp != websocket.TextMessage {
		k.Log("got wrong WS message", mtyp)

		switch mtyp {
		case websocket.CloseMessage:
			return nil, errors.New("received unexpected CloseMessage")

		case websocket.PingMessage, websocket.PongMessage:
			return nil, errors.New("received unexpected Ping/Pong message")

		default:
			return nil, fmt.Errorf("received unrecognized, unexpected message type %d", mtyp)
		}
	} else {
		// And this is where all my troubble begins, and ends.
		res, err := validateResponse(req.Method, msg)

		if err != nil {
			k.Log("error in validate", err)
			return nil, err
		} else {
			k.Log("received", string(msg), res)

			// Only Queries produce legible errors. The rest just kinda... does not. o.o
			// If it did throw an error, it'd be above.
			queryErrors := []error{}
			if req.Method == api.APIMethodQuery {
				errMsgs := res.Result.Get(`#(status!="OK")#.result`)
				if errMsgs.IsArray() {
					errMsgs.ForEach(func(_, value gjson.Result) bool {
						queryErrors = append(queryErrors, errors.New(value.String()))
						return true
					})
				}
			}

			k.Log("done", res, queryErrors)
			return res, errors.Join(queryErrors...)
		}
	}
	//panic("reached execObj(...) unexpectedly")
}

// Execute directly on the underlying WebSockets connection
func (con *SurrealConn) execRaw(sql string, args map[string]interface{}) (*api.Response, error) {
	k := con.k.Extend("execRaw")
	k.Log("start", sql, args)
	return con.execObj(con.Caller.CallQuery(sql, args))
}

func (con *SurrealConn) execWithArgs(sql string, args map[string]interface{}) (driver.Result, error) {
	k := con.k.Extend("execWithArgs")
	k.Log("start", sql, args)
	res, err := con.execObj(con.Caller.CallQuery(sql, args))
	if con.e.Debug(err) {
		return nil, err
	}
	newk := localKemba.Extend("result")
	return &SurrealResult{
		RawResult: res,
		k:         newk,
		e:         makeErrorLogger(newk),
	}, err
}

func (con *SurrealConn) queryWithArgs(sql string, args map[string]interface{}) (driver.Rows, error) {
	k := con.k.Extend("")
	k.Log("start", sql, args)
	res, err := con.execObj(con.Caller.CallQuery(sql, args))
	if con.e.Debug(err) {
		return nil, err
	}
	newk := localKemba.Extend("rows")
	return &SurrealRows{
		conn:      con,
		RawResult: res,
		resultIdx: 0,
		k:         newk,
		e:         makeErrorLogger(newk),
	}, err
}

func (con *SurrealConn) performLogin() error {
	k := con.k.Extend("performLogin")
	msg, err := con.Caller.CallSignin(con.creds)
	if con.e.Debug(err) {
		return err
	}
	rawMsg, _ := json.Marshal(msg)
	k.Log("message", string(rawMsg))
	res, err := con.execObj(msg)
	if con.e.Debug(err) {
		return err
	}
	// Attempt to run a `use [ns, db]`. Strings are empty (thus "null") by default.
	if con.creds.Method == config.AuthMethodDB || con.creds.Method == config.AuthMethodRoot {
		msg = con.Caller.CallUse(con.creds.Namespace, con.creds.Database)
		res, err = con.execObj(msg)
		if con.e.Debug(err) {
			return err
		}
	}
	k.Log("success", res)
	return nil
}

func (con *SurrealConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	k := con.k.Extend("PrepareContext")
	k.Log("start", ctx, query)
	// TODO: Take advantage of a given context.
	newk := localKemba.Extend("stmt")
	return &SurrealStmt{
		conn:  con,
		query: query,
		k:     newk,
		e:     makeErrorLogger(newk),
	}, nil
}
func (con *SurrealConn) Prepare(query string) (driver.Stmt, error) {
	k := con.k.Extend("Prepare")
	k.Log("start", query)
	return con.PrepareContext(context.Background(), query)
}

func (con *SurrealConn) Close() error {
	k := con.k.Extend("Close")
	k.Log("bye")
	return con.WSClient.Close()
}

func (con *SurrealConn) Begin() (driver.Tx, error) {
	k := con.k.Extend("Begin")
	k.Log("start")
	return con.BeginTx(context.Background(), driver.TxOptions{})
}

func (con *SurrealConn) IsValid() bool {
	k := con.k.Extend("IsValid")
	if err := con.WSClient.WriteMessage(websocket.PingMessage, nil); err != nil {
		k.Log("not valid", err)
		return false
	}
	k.Log("is still valid")
	return true
}

func (con *SurrealConn) ExecContext(ctx context.Context, sql string, args []driver.NamedValue) (driver.Result, error) {
	k := con.k.Extend("ExecContext")
	k.Log("start", ctx, sql, args)
	// TODO: Use the provided context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		k.Log("converting args")
		mappedValues := map[string]interface{}{}
		for _, v := range args {
			var key string
			if len(v.Name) == 0 || v.Name == "" {
				key = "_" + strconv.Itoa(v.Ordinal)
			} else {
				key = "_" + v.Name
			}
			k.Log("picking:", key, " O: ", v.Ordinal, " N: ", v.Name, " V: ", v.Value)
			mappedValues[key] = v.Value
		}
		return con.execWithArgs(sql, mappedValues)
	}
}

// QueryContext implements driver.QueryerContext.
func (con *SurrealConn) QueryContext(ctx context.Context, sql string, args []driver.NamedValue) (driver.Rows, error) {
	k := con.k.Extend("QueryContext")
	k.Log("start", ctx, sql, args)
	// TODO: Use the provided context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		k.Log("converting args")
		mappedValues := map[string]interface{}{}
		for _, v := range args {
			var key string
			if len(v.Name) == 0 || v.Name == "" {
				key = "_" + strconv.Itoa(v.Ordinal)
			} else {
				key = "_" + v.Name
			}
			k.Log("picking:", key, " O: ", v.Ordinal, " N: ", v.Name, " V: ", v.Value)
			mappedValues[key] = v.Value
		}
		return con.queryWithArgs(sql, mappedValues)
	}
}

func (con *SurrealConn) Exec(sql string, values []driver.Value) (driver.Result, error) {
	k := con.k.Extend("Exec")
	k.Log("start")
	mappedValues := map[string]interface{}{}
	for key, v := range values {
		mappedValues["_"+string(rune(key))] = v
	}
	return con.execWithArgs(sql, mappedValues)
}

// implements driver.ConnBeginTx
func (con *SurrealConn) BeginTx(ctx context.Context, _ driver.TxOptions) (driver.Tx, error) {
	k := con.k.Extend("BeginTx")
	k.Log("start", ctx)
	// TODO: Figure out how to take the provided ctx into account.
	// TODO: Can we use the TxOptions?
	if !con.IsValid() {
		return nil, driver.ErrBadConn
	}

	// TODO: Use the response to determine if everything is still fine.
	_, err := con.Exec("BEGIN TRANSACTION;", nil)
	if err != nil {
		return nil, err
	}

	return &SurrealConnBeginTx{
		conn: con,
	}, nil
}

func (con *SurrealConn) CheckNamedValue(nv *driver.NamedValue) (err error) {
	k := con.k.Extend("CheckNamedValue")
	k.Log("start", nv.Name, nv.Ordinal, nv.Value)
	nv.Value, err = checkNamedValue(nv.Value)
	return
}

func (con *SurrealConn) ConvertValue(v any) (driver.Value, error) {
	k := con.k.Extend("ConvertValue")
	k.Log("start", v)
	return checkNamedValue(v)
}

func (con *SurrealConn) Ping(ctx context.Context) error {
	k := con.k.Extend("Ping")
	k.Log("pingpongdong")
	// TODO: Figure out how to use contexts
	if !con.IsValid() {
		return errors.New("invalid connection")
	}
	msg, err := websocket.NewPreparedMessage(websocket.PingMessage, nil)
	if con.e.Debug(err) {
		return err
	}
	return con.WSClient.WritePreparedMessage(msg)
}
