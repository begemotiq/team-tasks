package mysql

import "database/sql"

func closeRows(rows *sql.Rows, err *error) {
	if closeErr := rows.Close(); *err == nil && closeErr != nil {
		*err = closeErr
	}
}
