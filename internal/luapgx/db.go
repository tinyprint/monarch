package luapgx

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	lua "github.com/yuin/gopher-lua"
)

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func dbExec(ctx context.Context, db Querier) func(*lua.LState) int {
	return func(L *lua.LState) int {
		sql := L.CheckString(1)
		paramsTable := L.OptTable(2, L.NewTable())

		var paramsSlice []interface{}
		paramsTable.ForEach(func(i lua.LValue, v lua.LValue) {
			paramsSlice = append(paramsSlice, v)
		})

		_, err := db.Exec(ctx, sql, paramsSlice...)
		if err != nil {
			L.RaiseError(err.Error())
			return 0
		}

		return 0
	}
}

func dbQuery(ctx context.Context, db Querier) func(*lua.LState) int {
	return func(L *lua.LState) int {
		sql := L.CheckString(1)
		paramsTable := L.OptTable(2, L.NewTable())

		var paramsSlice []interface{}
		paramsTable.ForEach(func(i lua.LValue, v lua.LValue) {
			paramsSlice = append(paramsSlice, v)
		})

		rows, err := db.Query(ctx, sql, paramsSlice...)
		if err != nil {
			defer rows.Close()
			L.RaiseError(err.Error())
			return 0
		}

		columns := rows.FieldDescriptions()
		columnCount := len(columns)

		resultTable := L.NewTable()

		L.SetField(resultTable, "rows", L.NewFunction(func(iter *lua.LState) int {
			if !rows.Next() {
				rows.Close()
				return 0
			}

			values, err := rows.Values()
			if err != nil {
				rows.Close()
				L.RaiseError(err.Error())
				return 0
			}

			rowTable := iter.CreateTable(columnCount, columnCount)
			for c, value := range values {
				column := columns[c]
				columnName := column.Name
				columnIndex := c + 1

				lVal, err := pgxToLuaValue(column.DataTypeOID, value)
				if isUnknownColumnTypeError(err) {
					rows.Close()
					return raiseUnknownColumnTypeError(iter, columnIndex, columnName, value)
				} else if err != nil {
					rows.Close()
					L.RaiseError(err.Error())
					return 0
				}

				if lVal != nil {
					iter.SetField(rowTable, columnName, lVal)
					rowTable.Insert(columnIndex, lVal)
				}
			}

			iter.Push(rowTable)

			return 1
		}))

		L.SetField(resultTable, "close", L.NewFunction(func(_ *lua.LState) int {
			rows.Close()
			return 0
		}))

		columnsTable := L.CreateTable(columnCount, 0)
		for i, column := range columns {
			columnsTable.Insert(i+1, lua.LString(column.Name))
		}
		L.SetField(resultTable, "columns", columnsTable)

		L.Push(resultTable)

		return 1
	}
}

func raiseUnknownColumnTypeError(L *lua.LState, colIndex int, colName string, colValue any) int {
	L.RaiseError("column %s (index %d) is of an unsupported type (%T); cast the value to a varchar or another type in your SQL query",
		colName, colIndex, colValue)
	return 0
}
