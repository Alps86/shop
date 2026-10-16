package hugo

import (
	"encoding/json"
	"errors"
	shophttp "github.com/Alps86/shop/internal/adapter/http"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

func New(snapshot application.AdminCatalogSnapshot, locale value.Locale) Pages {
	pages := Pages{}
	for _, product := range snapshot.Products {
		if product.Published {
			pages = append(pages, source{"product", string(product.ID), product.Slug, product.Name[locale], product.Description[locale], product.Revision, locale}.page())
		}
	}

	for _, category := range snapshot.Categories {
		pages = append(pages, source{"category", string(category.ID), category.Slug, category.Name[locale], category.Description[locale], category.Version, locale}.page())
	}

	return pages
}

func (input source) page() Page {
	segments := map[string]string{"product": "produkt", "category": "kategorie"}
	canonical := "/" + string(input.locale) + "/" + segments[input.kind] + "/" + input.slug + "--" + input.id

	return Page{Title: input.name, Description: input.description, URL: "/_frame/" + input.kind + "/" + input.id + "/" + strconv.FormatUint(input.version, 10) + "/", Layout: "dynamic", FrameKind: input.kind, RootVersion: input.version, Canonical: canonical, Fragment: "/fragments" + canonical, ID: input.id}
}

func (pages Pages) Routes() []shophttp.PublicRoute {
	routes := []shophttp.PublicRoute{{Path: "/de-DE/", Title: "Technik für deinen Alltag", Kind: "home"}, {Path: "/de-DE/service", Title: "Service und Kontakt", Kind: "editorial"}, {Path: "/de-DE/marken", Title: "Markenwelt", Kind: "editorial"}}
	for _, page := range pages {
		routes = append(routes, shophttp.PublicRoute{Path: page.Canonical, Title: page.Title, Description: page.Description, Kind: page.FrameKind})
	}

	return routes
}

func (page Page) Write(writer io.Writer) error {
	if err := json.NewEncoder(writer).Encode(page); err != nil {
		return err
	}

	_, err := io.WriteString(writer, "\n"+page.Description+"\n")

	return err
}

func (pages Pages) WriteDirectory(directory string) error {
	if err := os.MkdirAll(directory, 0750); err != nil {
		return err
	}

	for _, page := range pages {
		if err := page.writeFile(filepath.Join(directory, page.FrameKind+"-"+url.PathEscape(page.ID)+".md")); err != nil {
			return err
		}
	}

	return nil
}

func (page Page) writeFile(path string) (err error) {
	// #nosec G304 -- Zielpfad aus explizitem Exportverzeichnis und URL-escapierter Katalog-ID.
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	return page.Write(file)
}
