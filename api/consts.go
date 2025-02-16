package api

type APIMethod string

const (
	// # Auth
	APIMethodSignUp       APIMethod = "signup"       // -> string
	APIMethodSignIn       APIMethod = "signin"       // -> string
	APIMethodAuthenticate APIMethod = "authenticate" // -> ?
	APIMethodInvalidate   APIMethod = "invalidate"   // -> null

	// # Querying
	APIMethodQuery   APIMethod = "query"   // -> []{ status, time, result: []object }
	APIMethodGraphQL APIMethod = "graphql" // -> unique

	// # CRUD
	APIMethodSelect         APIMethod = "select"          // -> []object
	APIMethodCreate         APIMethod = "create"          // -> []object
	APIMethodInsert         APIMethod = "insert"          // -> []object
	APIMethodMerge          APIMethod = "merge"           // -> []object
	APIMethodPatch          APIMethod = "patch"           // -> JSON Patch
	APIMethodInsertRelation APIMethod = "insert_relation" // -> object{ in, out, id, ... }
	APIMethodRelate         APIMethod = "relate"          // -> object{ in, out, id, ... }
	APIMethodUpdate         APIMethod = "update"          // -> object
	APIMethodUpsert         APIMethod = "upsert"          // -> object
	APIMethodDelete         APIMethod = "delete"          // -> object

	// # Live Query
	APIMethodLive APIMethod = "live" // -> string (UUID)
	APIMethodKill APIMethod = "kill" // -> null

	// # Misc.
	APIMethodVersion APIMethod = "version" // -> string
	APIMethodInfo    APIMethod = "info"    // -> object{ ... }
	APIMethodUse     APIMethod = "use"     // -> null
	APIMethodLet     APIMethod = "let"     // -> null
	APIMethodUnset   APIMethod = "unset"   // -> null
	APIMethodRun     APIMethod = "run"     // -> any ()
)
