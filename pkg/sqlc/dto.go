package sqlc

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func NumericFromFloat(v float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(v)
	return n
}

func Int4FromInt32(v int32) pgtype.Int4 {
	var n pgtype.Int4
	_ = n.Scan(v)
	return n
}

func TimeToPgxTime(t time.Time) pgtype.Timestamp {
	var ts pgtype.Timestamp
	_ = ts.Scan(t)
	return ts
}
