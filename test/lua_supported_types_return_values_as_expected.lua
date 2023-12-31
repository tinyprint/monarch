db.exec("SET CLIENT_ENCODING TO 'UTF-8'", {})
db.exec("SET TIMEZONE TO 'America/Denver'", {})

local result = db.query([===[
    SELECT
        32767::smallint AS smallint,
        2147483647::integer AS integer,
        9223372036854775807::bigint AS bigint,
        1234567890.0987654321::decimal AS decimal,
        1234567890.0987654321::numeric AS numeric,
        3.4028235e38::real AS real,
        1.7976931348623157e308::double precision AS double,
        '52093.89'::money AS money,
        '550e8400-e29b-41d4-a716-446655440000'::uuid AS uuid,

        B'1010111100011'::bit(13)::varchar AS bit,
        B'1010111100011'::varbit::varchar AS varbit,
        true::boolean AS true,
        false::boolean AS false,
        null::boolean AS nullbool,
        'abc \153\154\155 \052\251\124'::bytea AS bytea,

        'characters'::char(10) AS char,
        'any amount of characters'::varchar AS varchar,
        'a lot of text'::text AS text,

        '2023-12-26'::date::varchar AS date,
        '5 years, 3 months, 12 days, 3 hours'::interval::varchar::varchar AS interval,
        '19:23:53'::time::varchar::varchar AS time,
        '19:23:53 EST'::timetz::varchar::varchar AS timetz,
        '2023-12-26 19:23:53'::timestamp::varchar AS timestamp,
        '2023-12-26 19:23:53 EST'::timestamptz::varchar AS timestamptz,

        '{"field": "value"}' AS json,
        '{"field": "value"}' AS jsonb
]===], {});

function octal2byte(octalAsString)
    return string.char(tonumber(octalAsString, 8))
end

local expected = {
    {name="smallint", value=32767},
    {name="integer", value=2147483647},
    {name="bigint", value=9223372036854775807},
    {name="decimal", value="1234567890.0987654321"},
    {name="numeric", value="1234567890.0987654321"},
    {name="real", value=3.4028234663852886e38},
    {name="double", value=1.7976931348623157e308},
    {name="money", value="$52,093.89"},
    {name="uuid", value="550e8400-e29b-41d4-a716-446655440000"},

    {name="bit", value="1010111100011"},
    {name="varbit", value="1010111100011"},
    {name="true", value=true},
    {name="false", value=false},
    {name="nullbool", value=null},
    {name="bytea", value='abc '
        ..octal2byte("153", 8)..octal2byte("154", 8)..octal2byte("155", 8)
        ..' '
        ..octal2byte("052", 8)..octal2byte("251", 8)..octal2byte("124", 8)},

    {name="char", value="characters"},
    {name="varchar", value="any amount of characters"},
    {name="text", value="a lot of text"},

    {name="date", value="2023-12-26"},
    {name="interval", value="5 years 3 mons 12 days 03:00:00"},
    {name="time", value="19:23:53"},
    {name="timetz", value="19:23:53-05"},
    {name="timestamp", value="2023-12-26 19:23:53"},
    {name="timestamptz", value="2023-12-26 17:23:53-07"},

    {name="json", value='{"field": "value"}'},
    {name="jsonb", value='{"field": "value"}'},
}

for i, expectedColumn in ipairs(expected) do
    if result.columns[i] ~= expectedColumn.name then
        error(string.format(
            "column name at index %d expected to be %q; got %q",
            i,
            expectedColumn.name,
            result.columns[i]
        ))
    end
end

if #result.columns ~= #expected then
    error(string.format(
        "query expected to return %d result.columns; got %d",
        #result.columns,
        #expected
    ))
end

for row in result.rows do
    for i = 1, #result.columns do
        if type(row[expected[i].name]) == "table" and type(expected[i].value) == "table" then
            actualT = row[expected[i].name]
            expectedT = expected[i].value
            for k, v in pairs(actualT) do
                if actualT[k] ~= expectedT[k] then
                    error(string.format(
                        "column %s.%s expected value to be %q; got %q",
                        expected[i].name,
                        k,
                        expectedT[k],
                        actualT[k]
                    ))
                end
            end
        elseif row[expected[i].name] ~= expected[i].value then
            error(string.format(
                "column %s expected value to be %q; got %q",
                expected[i].name,
                expected[i].value,
                row[expected[i].name]
            ))
        end

        if row[expected[i].name] ~= row[i] then
            error(string.format(
                "column %q and %d are expected to return the same value; column %q gave %q and column %d gave %q",
                expected[i].name,
                i,
                expected[i].name,
                row[expected[i].name],
                i,
                row[i]
            ))
        end
    end
end
