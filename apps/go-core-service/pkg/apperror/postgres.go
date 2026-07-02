package apperror

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgconn"
)

// PostgreSQL error codes
const (
	PgUniqueViolation     = "23505"
	PgForeignKeyViolation = "23503"
	PgCheckViolation      = "23514"
)

// WrapDBError converts common database errors into the appropriate AppError.
//
// Mapping:
//   - sql.ErrNoRows          → 404 Not Found
//   - PG unique violation     → 409 Conflict
//   - PG foreign key violation→ 400 Bad Request
//   - everything else         → 500 Internal (hides raw SQL details)
func WrapDBError(err error, resourceName string) *AppError {
	if err == nil {
		return nil
	}

	// sql.ErrNoRows → Not Found
	if errors.Is(err, sql.ErrNoRows) {
		return NewNotFound(resourceName)
	}

	// PostgreSQL-specific errors via pgx driver
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case PgUniqueViolation:
			return Wrap(err, NewConflict(resourceName+" already exists"))
		case PgForeignKeyViolation:
			return Wrap(err, NewBadRequest("invalid reference for "+resourceName))
		case PgCheckViolation:
			return Wrap(err, NewValidation("check constraint violated for "+resourceName))
		}
	}

	// Unknown / connection errors → hide SQL detail, wrap for logging
	return Wrap(err, NewInternal("database error"))
}
