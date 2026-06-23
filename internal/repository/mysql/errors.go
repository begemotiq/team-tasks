package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"task-service/internal/domain"
	"task-service/internal/metrics"
)

func mapSQLError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func mapMySQLError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return domain.ErrConflict
		case 1451, 1452:
			return fmt.Errorf("%w: foreign key violation", domain.ErrInvalidInput)
		}
	}
	return err
}

func recordDBError(repository, operation string, err error) {
	metrics.RecordDBError(repository, operation, err)
}
