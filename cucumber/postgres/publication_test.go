package postgres_test

import (
	"context"
	"github.com/Alps86/shop/internal/domain/catalog"
	"strconv"
	"time"

	"github.com/cucumber/godog"
)

func (current *world) publicationSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^lautet der weiterhin maßgebliche Datensatz$`, current.publicationObserved)
	ctx.Step(`^bleibt die importierte Herkunft dauerhaft nachvollziehbar$`, current.provenanceObserved)
	ctx.Step(`^unverifizierte öffentliche Referenzdaten ohne Pflichtangabenergänzung übernommen werden$`, current.sourcePending)
}

func (current *world) publicationObserved(input *godog.Table) error {
	snapshot, err := current.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	var version string
	err = current.admin.QueryRow(context.Background(), "SELECT payload->>'Version' FROM "+current.schema+".records WHERE kind='active-ruleset'").Scan(&version)
	if err != nil {
		return err
	}

	actual := map[string]string{"ActiveDatasetVersion": version, "Products": strconv.Itoa(len(snapshot.Products)), "Price": strconv.FormatInt(snapshot.Inventory[0].Price.MinorUnits, 10), "Stock": strconv.FormatInt(snapshot.Inventory[0].StockQuantity, 10)}

	return check(table(input), actual)
}

func (current *world) sourcePending(input *godog.Table) error {
	row := table(input)
	current.manifest.SourceURL = row["SourceURL"]
	parsed, err := time.Parse(time.RFC3339, row["RetrievedAt"])
	if err != nil {
		return err
	}

	current.manifest.RetrievedAt = parsed
	current.manifest.LicenseNote = row["LicenseNote"]
	current.manifest.Products[0].Manufacturer = catalog.EconomicOperator{}
	current.manifest.Products[0].Compliance.Safety = catalog.SafetyInformation{}
	current.manifest.Products[0].Compliance.Context.VerificationStatus = row["VerificationStatus"]

	return nil
}

func (current *world) provenanceObserved(input *godog.Table) error {
	var source, license string
	var retrieved time.Time
	err := current.admin.QueryRow(context.Background(), "SELECT source_url,retrieved_at,license_note FROM "+current.schema+".datasets WHERE version=$1", current.manifest.DatasetVersion).Scan(&source, &retrieved, &license)
	if err != nil {
		return err
	}

	return check(table(input), map[string]string{"SourceURL": source, "RetrievedAt": retrieved.UTC().Format(time.RFC3339), "LicenseNote": license})
}
