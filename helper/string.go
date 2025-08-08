package helper

import "database/sql"

func ChooseString(newVal, oldVal string) string {
	if newVal != "" {
		return newVal
	}
	return oldVal
}

func ChooseNullString(newVal string, oldVal sql.NullString) sql.NullString {
	if newVal != "" {
		return sql.NullString{String: newVal, Valid: true}
	}
	return oldVal
}
