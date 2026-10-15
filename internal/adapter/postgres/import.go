package postgres

import (
	"context"
	"strings"

	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (store *Store) Import(ctx context.Context, manifest ImportManifest) (ImportResult, error) {
	if err := manifest.Validate(); err != nil {
		return ImportResult{}, err
	}

	result := ImportResult{Code: "OK", DatasetVersion: manifest.DatasetVersion, ProductCount: len(manifest.Products)}
	err := store.transact(ctx, func(tx pgx.Tx) error {
		return manifest.publish(ctx, store, tx, &result)
	})

	return result, err
}

func (manifest ImportManifest) Validate() error {
	if value.IdentifierSyntax(manifest.DatasetVersion).Validate() != nil {
		return value.NewError("InvalidInput", "DatasetVersion")
	}

	catalog := catalog.NewCatalog()
	if err := manifest.configure(&catalog); err != nil {
		return err
	}

	if err := manifest.validateProducts(catalog); err != nil {
		return err
	}

	return manifest.validatePublication()
}

func (manifest ImportManifest) validateProducts(config catalog.Catalog) error {
	for _, product := range manifest.Products {
		config.Products[product.ID] = product
	}

	for _, product := range manifest.Products {
		next, err := config.Publish(product)
		if err != nil {
			return err
		}

		if err := manifest.validateDocuments(product); err != nil {
			return err
		}

		config = next
	}

	return nil
}

func (manifest ImportManifest) validateInventory() error {
	for _, inventory := range manifest.Inventories {
		if err := inventory.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (manifest ImportManifest) publish(ctx context.Context, store *Store, tx pgx.Tx, result *ImportResult) error {
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(741902)"); err != nil {
		return err
	}

	digest, err := manifest.digest(ctx, tx)
	if err != nil {
		return err
	}

	previous, err := manifest.previous(ctx, tx, digest)
	if err != nil {
		return err
	}

	if previous {
		result.Code = "Noop"
		return nil
	}

	return manifest.insert(ctx, store, tx, digest)
}

func (manifest ImportManifest) insert(ctx context.Context, store *Store, tx pgx.Tx, digest string) error {
	if err := manifest.insertConfiguration(ctx, tx); err != nil {
		return err
	}

	if err := manifest.insertProducts(ctx, store, tx); err != nil {
		return err
	}

	_, err := tx.Exec(ctx, "INSERT INTO datasets(version,digest,source_url,retrieved_at,license_note) VALUES($1,$2,$3,$4,$5)", manifest.DatasetVersion, digest, manifest.SourceURL, manifest.RetrievedAt, manifest.LicenseNote)

	return databaseError(err)
}

func (manifest ImportManifest) insertConfiguration(ctx context.Context, tx pgx.Tx) error {
	for _, family := range manifest.Families {
		if err := newRecord("family", string(family.ID), family).save(ctx, tx); err != nil {
			return err
		}
	}

	for _, category := range manifest.Categories {
		if err := newRecord("category", string(category.ID), category).save(ctx, tx); err != nil {
			return err
		}
	}

	return manifest.insertRules(ctx, tx)
}

func (manifest ImportManifest) insertRules(ctx context.Context, tx pgx.Tx) error {
	for _, rule := range manifest.DiscountRules {
		if err := newRecord("discount", string(rule.ID), rule).save(ctx, tx); err != nil {
			return err
		}
	}

	if err := newRecord("ruleset", manifest.DatasetVersion, manifest.DiscountRules).save(ctx, tx); err != nil {
		return err
	}

	if err := newRecord("active-ruleset", "active", versionState{Version: manifest.DatasetVersion}).save(ctx, tx); err != nil {
		return err
	}

	return newRecord("statistics", "synthetic", manifest.CoPurchaseStatistics).save(ctx, tx)
}

func (manifest ImportManifest) insertProducts(ctx context.Context, store *Store, tx pgx.Tx) error {
	if err := manifest.insertInventories(ctx, tx); err != nil {
		return err
	}

	for _, product := range manifest.Products {
		if err := store.publishProduct(ctx, tx, product); err != nil {
			return err
		}
	}

	return nil
}

func (manifest ImportManifest) insertInventories(ctx context.Context, tx pgx.Tx) error {
	for _, inventory := range manifest.Inventories {
		if _, err := tx.Exec(ctx, "INSERT INTO records(kind,id,payload) VALUES('inventory',$1,$2) ON CONFLICT DO NOTHING", inventory.SKU, inventory); err != nil {
			return err
		}
	}

	return nil
}

func (store *Store) publishProduct(ctx context.Context, tx pgx.Tx, product catalog.Product) error {
	if _, err := tx.Exec(ctx, "INSERT INTO projection_versions(product_id,revision) VALUES($1,0) ON CONFLICT(product_id) DO UPDATE SET deleted=false", product.ID); err != nil {
		return err
	}

	product, err := store.nextProduct(ctx, tx, product)
	if err != nil {
		return err
	}

	if err := store.registerSKUs(ctx, tx, product); err != nil {
		return err
	}

	if err := newRecord("product", string(product.ID), product).save(ctx, tx); err != nil {
		return err
	}

	return store.project(ctx, tx, product.ID)
}

func (store *Store) registerSKUs(ctx context.Context, tx pgx.Tx, product catalog.Product) error {
	for _, variant := range product.Variants {
		var owner string
		if err := tx.QueryRow(ctx, "INSERT INTO sku_registry(sku,product_id) VALUES($1,$2) ON CONFLICT(sku) DO UPDATE SET sku=EXCLUDED.sku RETURNING product_id", variant.SKU, product.ID).Scan(&owner); err != nil {
			return err
		}

		if strings.Compare(owner, string(product.ID)) != 0 {
			return value.NewError("DuplicateSKU", "SKU")
		}
	}

	return nil
}

func (store *Store) nextProduct(ctx context.Context, tx pgx.Tx, product catalog.Product) (catalog.Product, error) {
	var previous uint64
	err := tx.QueryRow(ctx, "SELECT COALESCE((SELECT (payload->>'Revision')::bigint FROM records WHERE kind='product' AND id=$1),0)", product.ID).Scan(&previous)
	product.Revision = previous + 1
	product.Published = true

	return product, databaseError(err)
}
