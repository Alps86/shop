package postgres

import (
	"context"

	"github.com/Alps86/shop/internal/domain/value"
)

func (command commandRecord) previous(ctx context.Context, db database) (bool, error) {
	var matches bool
	err := db.QueryRow(ctx, "SELECT payload=$3::jsonb,result FROM commands WHERE scope=$1 AND command_id=$2", command.scope, command.id, command.payload).Scan(&matches, command.result)
	if value.Code(databaseError(err)) == "NotFound" {
		return false, nil
	}

	if err != nil {
		return false, databaseError(err)
	}

	if !matches {
		return true, value.NewError("Conflict", "CommandID")
	}

	return true, nil
}

func (command commandRecord) save(ctx context.Context, db database) error {
	_, err := db.Exec(ctx, "INSERT INTO commands(scope,command_id,payload,result) VALUES($1,$2,$3,$4)", command.scope, command.id, command.payload, command.result)
	return databaseError(err)
}
