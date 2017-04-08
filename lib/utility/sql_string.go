package utility

import "database/sql"

func NullStringCreate(str string) sql.NullString {
	if str == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{Valid: true, String: str}
}
func NullStringGet(str sql.NullString) string {
	if str.Valid {
		return str.String
	}
	return ""
}
