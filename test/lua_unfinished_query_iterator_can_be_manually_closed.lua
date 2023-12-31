local result = db.query([===[
    SELECT 1
    UNION
    SELECT 2
]===], {});

for row in result.rows do
    -- retrieve only one row to leave `rows` unclosed
    break
end
result.close()

local moreRows = db.query([===[
    SELECT 3
]===], {});
