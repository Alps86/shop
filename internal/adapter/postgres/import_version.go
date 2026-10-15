package postgres

import (
	"context"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (manifest ImportManifest) digest(ctx context.Context, db database) (string, error) {
	var digest string
	err := db.QueryRow(ctx, "SELECT encode(sha256(convert_to($1::jsonb::text,'UTF8')),'hex')", manifest).Scan(&digest)

	return digest, databaseError(err)
}

func (manifest ImportManifest) previous(ctx context.Context, tx pgx.Tx, digest string) (bool, error) {
	var existing string
	err := tx.QueryRow(ctx, "SELECT digest FROM datasets WHERE version=$1", manifest.DatasetVersion).Scan(&existing)
	if value.Code(databaseError(err)) == "NotFound" {
		return false, nil
	}

	if err != nil {
		return false, databaseError(err)
	}

	if existing != digest {
		return false, value.NewError("Conflict", "DatasetVersion")
	}

	return true, nil
}
