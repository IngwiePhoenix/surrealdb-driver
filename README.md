# SurrealDB <-> Go: The only `database/sql` compliant SurrealDB driver (so far)!

> [!IMPORTANT]  
> This driver is not entirely stable just yet. Beware, there do be dragons amongst some type-asserts. ;)

This is a "about a week" project; implementing Go's `database/sql` (and `database/sql/driver`) interfaces in order to allow me to use any kind of Go-tooling I could want.

## But, why?

Most of my code these days is written in Go. I feel confortable with the language and it's semantics. But, I have worked with MySQL and Postgres so much that I just want to use something different and new. Also, I couldn't write a `JOIN` to save my life, to be real honest... And SurrealDB solves that - by just not having `JOIN`s, at all! In fact, it even does related data resolution and handles n:m like it's _nothing_.

## How to use it?

It's your bog-standard `database/sql` interface - and thus, quite easy.

1. Blank-import it: `import _ "github.com/IngwiePhoenix/surrealdb-driver`
2. Create a new DB connection: `db, err := sql.Open("surrealdb", "ws://user:pass@host:port/rpc?method=root&db=yourdb&ns=yourdb")
   - **`ws://...` or `wss://...`**: SurrealDB works over WebSockets. The pure HTTP interface is not (directly) supported by this driver.
   - **`host:port/rpc`**: Aside from specifying your hostname/IP and port, you also must specify the path to the RPC endpoint. This allows you to use a SurrealDB instance ran in a sub-path. For instance: `wss://user:pass@somecloudinstance.org/instanceid/rpc?...`
   - **Query parameters**: This driver does not reinvent the wheel - it's just a URL, including creds. This URL however is parsed and taken apart for it's data - your username and password (and other parameters) are never sent verbatim - they will only be sent _after_ a successful connection.
     - **`method=`**: One of: `root`, `db` or `record`.
       - `method=root`: Log in with username and password as a "root user". You may optionally specify `ns=` and `db=` and the driver will select those upon authentication.
       - `method=db`: Signin with database-level permissions. This requires username and password as well as `ns=` and `db=` to be set.
       - `method=record`: This uses the record authentication and goes one level deeper than `method=db`; you also need to supply `ac=`.
       - `method=token`: This is currently not implemented (Token authentication)
       - `method=anon`: This is currently not implemented (Anonymous authentication)
     - **`ns=...`**: Specify your desired namespace.
     - **`db=...`**: Specify your desired database.
     - **`ac=...`**: Specify your access control name.
3. Make queries! `rows, err := db.Query("SELECT * FROM users;")`
   - ...and use SurrealDB features. This driver sends the query straight to SurrealDB with nearly no pre- or post-processing. Consult the [SurrealQL](https://surrealdb.com/docs/surrealql) documentation for more inforation!
   - ...and perhaps make multiples. This driver supports returning everything, even the result of multiple queries, incorporates the errors into the `.Err()` method and more!
   - ...or, cast it, and use it raw and directly (very advanced): `db.Conn().(*surrealdbdriver.SurrealConn)`.
     - This will give you access to the `.Caller` field, which can construct WebSocket requests for you, and `.WSClient` with which you can send them.
     - Be aware that this is part of Go's methods and you must adhere to their rules of closing an obtained connection properly.

### Using the `rel` adapter

This is pretty straight forward:

```go
package main
import (
	"log"

    // Import the driver itself to register it
	_ "github.com/IngwiePhoenix/surrealdb-driver"

    // This imports the rel adapter
	srel "github.com/IngwiePhoenix/surrealdb-driver/pkg/rel"

    // Import rel itself for the ORM API
	"github.com/go-rel/rel"
)
func main() {
	adapter, err := srel.Open("ws://root:root@127.0.0.1:8000/rpc?method=root")
	if err != nil {
		log.Println("Could not make adapter")
		log.Fatal(err.Error())
	}
	defer adapter.Close()

	repo := rel.New(adapter)
}
```

You can now use the [official `rel` documentation](https://go-rel.github.io/) to interface with SurrealDB.

A few nicities:

- When using migrations, you may use `t.Column(...)` to specify a custom data type, such as `record<something>`.
- Everything is optional by default; use `rel.Required(true)` to change that. This is equivalent of setting `DEFAULT NOT NULL` in most other DBMS. Otherwise, your designated type is wrapped in `option<T>`.
- There is currently very limited support for indexes; SurrealDB has a lot of custom index kinds (search index and alike) so this will take me a moment.
- You currently can not control concurrentness (`PARALLEL` or `CONCURRENT`) in statements.
- As SurrealDB lacks renaming, it will **copy everything** on field or table renaming. Be aware of that. Also, this will trigger implicit casting of your current data. Rather than renaming, you should instead create a new field yourself and copy the data manually, while applying casts. Use a raw SQL fragment to do so. A little hacky, but considerably nieche as it is.

## Tool integrations

I am currently working on integrating with these amazing tools:

- [GORM](https://gorm.io/): An implementation for that is underway in `pkg/gorm`. However, it's interface is rather complicated and nested, so I have focused elsewhere.
- [rel](https://github.com/go-rel): This is working, and has migrations. Genuenly recommend this. It can be found in `pkg/rel` - however, I might have to rename that to avoid name clashing.
- [golang-migrate/migrate](https://github.com/golang-migrate/migrate): Considering to implement this outside of `rel`, if only to learn more about the internals.
- [xo/usql](https://github.com/xo/usql): The last database client you might ever need. Connects to practically everything - and soon, to SurrealDB too! To implement it, `xo/dburl` must be implemented for this first, [which I have learned](https://github.com/xo/dburl/issues/43). Due to the architecture of the project, I will need substantially more testing before I finish this part - I need a very stable driver.

While SurrealDB itself is fantastic and has many great features, it lacks proper tooling. Yes, there a few in [awesome-surrealdb](https://github.com/surrealdb/awesome-surreal) but most of them are in JavaScript and TypeScript - not ideal for every situation. For example, I plan to write an app with [templ](https://templ.guide/)...so I won't even include a JavaScript runtime in my DevContainer - because, I am not gonna need one. Hence, having standalone clients and tools just makes things a wee bit easier. :) Who can say "no" to a little bit of DX (Developer Experience)?

But overall, I just hope it helps someone adopt SurrealDB!

## What about their official `surrealdb.go` client?

This is a standalone client, that _can not_ unmarshal into custom structs, especially if you need related and ascending data. Heck, [it can't even query it's own version...](https://github.com/surrealdb/surrealdb.go/issues/183) So, I wrote a "more generic" client - it does not use CBOR, so you won't need to implement `(Un)marshalCBOR` for your own types and works with standard `json:"..."` tags. However, it _might_ actually be slower... But, that's a different debate.

## What I learned...

Having a predetermined, specified, strict API scheme can save you time and headache killers. A **lot** of them.
