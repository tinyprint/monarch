package luapgx

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	lua "github.com/yuin/gopher-lua"
)

type errUnknownColumnType struct {
	error
	value any
}

func (e *errUnknownColumnType) unknownColumnType() string {
	return fmt.Sprintf("%T", e.value)
}

func (e *errUnknownColumnType) Error() string {
	return fmt.Sprintf("unknown column type %T", e.value)
}

func isUnknownColumnTypeError(err error) bool {
	typedErr := &errUnknownColumnType{}
	isUnknownColumnType := errors.As(err, &typedErr)
	return isUnknownColumnType
}

func pgxToLuaValue(dataTypeOID uint32, value any) (lua.LValue, error) {
	var lVal lua.LValue

	switch val := value.(type) {
	case string:
		lVal = lua.LString(val)
	case [16]byte: // uuid
		if pgtype.UUIDOID == dataTypeOID {
			lVal = lua.LString(fmt.Sprintf(
				"%x-%x-%x-%x-%x",
				val[0:4],
				val[4:6],
				val[6:8],
				val[8:10],
				val[10:16],
			))
		} else {
			return nil, &errUnknownColumnType{value: value}
		}
	case []byte:
		lVal = lua.LString(val)
	//case time.Time:
	//	timeTable := L.NewTable()
	//
	//	if val.Location() != time.UTC {
	//		return nil, fmt.Errorf("lua does not support timezones; convert column to a string, date, time, or timestamp (without tz)")
	//	}
	//
	//	L.SetField(timeTable, "year", lua.LNumber(val.Year()))
	//	L.SetField(timeTable, "month", lua.LNumber(val.Month()))
	//	L.SetField(timeTable, "day", lua.LNumber(val.Day()))
	//	L.SetField(timeTable, "hour", lua.LNumber(val.Hour()))
	//	L.SetField(timeTable, "min", lua.LNumber(val.Minute()))
	//	L.SetField(timeTable, "sec", lua.LNumber(val.Second()))
	//	L.SetField(timeTable, "isdst", lua.LBool(val.IsDST()))
	//	L.SetField(timeTable, "yday", lua.LNumber(val.YearDay()))
	//	// in Go, Sunday=0 ... Saturday=6; in Lua, Sunday=1 ... Saturday=7
	//	L.SetField(timeTable, "wday", lua.LNumber(val.Weekday()+1))
	//
	//	if pgtype.DateOID == dataTypeOID {
	//		L.SetField(timeTable, "sql", lua.LString(val.Format(time.DateOnly)))
	//	} else {
	//		L.SetField(timeTable, "sql", lua.LString(val.Format(time.RFC3339)))
	//	}
	//
	//	lVal = timeTable
	//case pgtype.Interval:
	//	intervalTable := L.NewTable()
	//
	//	L.SetField(intervalTable, "months", lua.LNumber(val.Months))
	//	L.SetField(intervalTable, "days", lua.LNumber(val.Days))
	//	L.SetField(intervalTable, "microseconds", lua.LNumber(val.Microseconds))
	//
	//	lVal = intervalTable
	//case pgtype.Time:
	//	timeTable := L.NewTable()
	//
	//	fmt.Println(dataTypeOID)
	//	L.SetField(timeTable, "microseconds", lua.LNumber(val.Microseconds))
	//
	//	lVal = timeTable
	case pgtype.Numeric:
		asString, err := val.MarshalJSON()
		if err != nil {
			return nil, err
		}
		lVal = lua.LString(asString)
	//case pgtype.Bits:
	//	bitsTable := L.NewTable()
	//
	//	var number uint = 0
	//	for i, bits := range val.Bytes {
	//		// if we already placed a byte worth of bits on the number,
	//		// we need to shift all bits left a byte to make way for the
	//		// next byte worth of bits
	//		if i > 0 {
	//			number = number << 8
	//		}
	//
	//		// add the set of bits to the number
	//		number = number | uint(bits)
	//	}
	//
	//	// when the bit string length is not a multiple of 8,
	//	// we need to strip off the extra digits on the right
	//	partialByteSize := val.Len % 8
	//	if partialByteSize != 0 {
	//		number = number >> (8 - partialByteSize)
	//	}
	//
	//	L.SetField(bitsTable, "number", lua.LNumber(number))
	//	L.SetField(bitsTable, "string", lua.LString(fmt.Sprintf("%b", number)))
	//
	//	lVal = bitsTable
	case bool:
		lVal = lua.LBool(val)
	case int:
		lVal = lua.LNumber(val)
	case int16:
		lVal = lua.LNumber(val)
	case int32:
		lVal = lua.LNumber(val)
	case int64:
		lVal = lua.LNumber(val)
	case float32:
		lVal = lua.LNumber(val)
	case float64:
		lVal = lua.LNumber(val)
	//case fmt.Stringer:
	//	lVal = lua.LString(val.String())
	case nil:
		lVal = lua.LNil
	default:
		return nil, &errUnknownColumnType{value: value}
	}

	return lVal, nil
}
