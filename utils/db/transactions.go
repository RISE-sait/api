package db

import (
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
)

// ExecuteInTx wraps a function in a database transaction, handling the begin/commit/rollback cycle.
// It takes a context, a database connection, and a function that operates within the transaction.
// The provided function receives a transaction object and should return a CommonError if any error occurs.
//
// If the transaction begins successfully, the function:
// 1. Executes the provided function with the transaction
// 2. Commits the transaction if the function succeeds
// 3. Automatically rolls back the transaction if either the function fails or commit fails
//
// Parameters:
//   - ctx: Context for the transaction
//   - db: Database connection to use
//   - fn: Function to execute within the transaction
//
// Returns:
//   - *CommonError: nil if successful, error details if any operation fails
//
// Example:
//
//	err := ExecuteInTx(ctx, db, func(tx *sql.Tx) *errLib.CommonError {
//	    // Perform database operations using tx
//	    return nil
//	})
func ExecuteInTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) *errLib.CommonError) *errLib.CommonError {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("Rollback error (usually harmless): %v", err)
		}
	}()

	if txErr := fn(tx); txErr != nil {
		return txErr
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	return nil
}
