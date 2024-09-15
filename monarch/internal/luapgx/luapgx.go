package luapgx

import (
	"context"

	lua "github.com/yuin/gopher-lua"
)

func NewDBTable(ctx context.Context, L *lua.LState, globalName string, db Querier) {
	table := L.NewTable()
	L.SetFuncs(table, map[string]lua.LGFunction{
		"exec":  dbExec(ctx, db),
		"query": dbQuery(ctx, db),
	})
	L.SetGlobal(globalName, table)
}
