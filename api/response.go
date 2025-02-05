package api

import (
	"errors"
	"strconv"

	st "github.com/senpro-it/dsb-tool/extras/surrealdb-driver/surrealtypes"
)

type ErrorDetails struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ErrorDetails) ToError() error {
	c := strconv.Itoa(e.Code)
	m := e.Message
	return errors.New(c + ": " + m)
}

type GenericResponse[T ResultTypes] struct {
	// ID picked by the requester, returned in kind
	Id RequestID `json:"id,omitempty"`
	// May or may not exist; contains higher level error information
	Error *ErrorDetails `json:"error"`
	// Completely variadic. WTF?
	// - On Query or alike: One or more of T
	// - On methods like "info": null or T
	// ...and so forth. Why?!
	// ref: https://github.com/surrealdb/surrealdb/blob/main/crates/core/src/rpc/response.rs
	Result *T `json:"result"`
}

type LiveNotificationResponse struct {
	Action string `json:"action"`
	Id string `json:"id"`
	Result st.Object `json:"result"`
}

// A response with no content but an actual error.
// Returned by everything, eventually, potentially.
// .error is ONLY set IF an error exists/has occured
type FatalErrorResponse = GenericResponse[interface{}]

// Response to the query command.
// Guaranteed to always be an array (Rust Vec<(Database)::Value>)
// this represents the direct (Surreal)SQL interface.
type QueryResponse = GenericResponse[[]QueryResult]

/*
Select: array of objects
Create: array of objects
Insert: array of objects
Merge: array of objects
Insert Relation: Object
Update: object
Upsert: object
Delete: object
Relate: object
*/


// Response to these commands:
//   select, create, insert, merge
// Contains one object for each affected entry.
// This object IS NOT the same as a QueryResult; it's missing "header" data.
// It should be an array response if multiple values are affected.
// Unfortunately, there is no easy tell... do a `v, ok := resp.([]interface{})`
// type assert, and hope for the best. I'm just going off of the docs.
type MultiNoSQLResponse = GenericResponse[[]st.Object]

// Response to:
//   update, upsert, delete
//   (relate, insert_relation too but they are handled elsewhere)
// The same principals apply as above.
type SingleNoSQLResponse = GenericResponse[st.Object]

// Response to the live command, contains a query UUID
type LiveResponse = GenericResponse[string]

// Response to the kill command, contains nothing.
type KillResponse = GenericResponse[interface{}]]

// Response to the run command, can contain anything.
// Output is dependant on the function.
// It's probably better to use GenericResponse[T] yourself?
type RunResponse = GenericResponse[any]

// Response to the graphql command, contains specialized information.
// Highly implementation specific; generalized here.
type GraphQLResponse = GenericResponse[st.Object]

// Response to the insert and insert_relation command, returns exactly one relation.
type RelationResponse = GenericResponse[RelationResult]

// Response to the patch command and returns an array of operations.
// We use a dependency here that knows how to deal with that.
type PatchResponse = GenericResponse[JsonPatchResult]

// Response to the version command.
// The docs say it's an object, but in my tests, it was a string.
// ...so now it's a string, I'sppose.
type VersionResponse = GenericResponse[string]

// Response to:
//   signin, signup,
//   authenticate, invalidate
// Contains a string or null
type AuthResponse = GenericResponse[string]

// Response to the info command, may contain an object.
type InfoResponse = GenericResponse[st.Object]

// Response to let and unset. Always null or an error.
type VarRequest = GenericResponse[interface{}]