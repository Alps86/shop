package runtime

import (
	"context"
	"encoding/json"
	"os"
)

func (app *App) importSeed(ctx context.Context) error {
	if app.Config.SeedFile == "" {
		return nil
	}

	payload, err := os.ReadFile(app.Config.SeedFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(payload, &app.Seed); err != nil {
		return err
	}

	_, err = app.Store.Import(ctx, app.Seed.ImportManifest)

	return err
}
