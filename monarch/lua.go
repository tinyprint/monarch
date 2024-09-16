package monarch

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/tinyprint/monarch/monarch/internal/luapgx"
)

type runLuaConfig struct {
	file string
}

func luaEnv() (*lua.LState, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})

	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.MathLibName, lua.OpenMath},
		{lua.StringLibName, lua.OpenString},
		{lua.IoLibName, lua.OpenIo},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			return nil, err
		}
	}

	return L, nil
}

func runLua(ctx context.Context, db luapgx.Querier, config runLuaConfig) error {
	L, err := luaEnv()
	if err != nil {
		return err
	}

	luapgx.NewDBTable(ctx, L, "db", db)

	if err := L.DoFile(config.file); err != nil {
		return err
	}

	return nil
}
