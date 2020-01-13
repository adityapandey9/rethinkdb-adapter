RethinkDB Adapter 
====

RethinkDB Adapter is the [Rethink DB](https://www.rethinkdb.com/) adapter for [Casbin](https://github.com/casbin/casbin). With this library, Casbin can load policy from RethinkDB or save policy to it.

## Installation

    go get github.com/adityapandey9/rethinkdb-adapter

## Simple Example

```go
package main

import (
    	"os"
    	r "gopkg.in/gorethink/gorethink.v3"
	"github.com/casbin/casbin"
	"github.com/adityapandey9/rethinkdb-adapter"
)

func getConnect() r.QueryExecutor {
	url := os.Getenv("RETHINKDB_URL") //Get the Rethinkdb url from system env

	if url == "" {
		url = "localhost:28015"
	}

	session, _ := r.Connect(r.ConnectOpts{
		Address: url,
	})

	return session
}

func main() {
	// Initialize a RethinkDB get session, add it to adapter and use it in a Casbin enforcer:
	// The adapter will use the database named "casbin".
	// If it doesn't exist, the adapter will create it automatically. (default names - Database: casbin, Table: rethinkdbpolicy)
  	session := getConnect()
	a := rethinkadapter.NewAdapter(session) // Your RethinkDB Session. 
	//Or you can do this
	a := rethinkadapter.NewAdapterDB(session, "database_name", "table_name") // Your RethinkDB Session.
	
	e := casbin.NewEnforcer("examples/casbinmodel.conf", a)
	
	// Load the policy from DB.
	e.LoadPolicy()
	
	// Check the permission.
	e.Enforce("alice", "data1", "read")
	
	// Modify the policy.
	// e.AddPolicy(...)
	// e.RemovePolicy(...)
	
	// Save the policy back to DB.
	e.SavePolicy()
}
```

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.
