package api

import (
	"errors"
	"strconv"

	"github.com/tidwall/gjson"
)

// implements error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var _ (error) = (*APIError)(nil)

func (e *APIError) String() string {
	c := strconv.Itoa(e.Code)
	m := e.Message
	return "surrealdb: " + c + " - " + m
}

func (e *APIError) Error() string {
	return e.String()
}

func (e *APIError) ToError() error {
	return errors.New(e.String())
}

type Response struct {
	Method APIMethod
	Result gjson.Result
}

type LiveNotificationResponse struct {
	Action string       `json:"action"`
	Id     string       `json:"id"`
	Result gjson.Result `json:"result"`
}

/*
TL;DR:
	# Auth
	Signin:				string (token)
	Signup:				string (token)
	Invalidate:			null
	Info:				object

	# Querying
	Query:				object + array of objects (special)
	GraphQL:			object (special)

	# CRUD / NoSQL-ish
	Select: 			array of objects
	Create: 			array of objects
	Insert: 			array of objects
	Merge: 				array of objects
	Patch:				array of objects (JSON Patch)
	Insert Relation: 	Object
	Relate: 			object
	Update: 			object
	Upsert: 			object
	Delete: 			object

	# Live queries
	Live:				string
	Kill:				null

	# Misc.
	Version:			string
	Use:				null
	Let:				null
	Unset:				null
	Run:				literally anything

If any of them fail, the reply has an .error property - and thus,
has it's own response type with no .result

This inconsistency is gonna spell my doom. -.-
*/
