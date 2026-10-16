package http_test

import (
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"os"
	"strconv"
)

func (engine *compositionEngine) prepareSource(fixture fixture) error {
	engine.source = engine.directory + "/source"
	if err := os.CopyFS(engine.source, os.DirFS("../../web")); err != nil {
		return err
	}

	product := fixture.Detail.Product
	category := fixture.Category.Category
	seeds := []frameSeed{{Kind: "product", ID: string(product.ID), Revision: product.Revision, Path: "/de-DE/produkt/" + product.Slug + "--" + string(product.ID), Title: product.Name["de-DE"], Description: product.Description["de-DE"], Markdown: fixture.FrameContent}, {Kind: "category", ID: string(category.ID), Revision: category.Version, Path: "/de-DE/kategorie/" + category.Slug + "--" + string(category.ID), Title: category.Name["de-DE"], Description: category.Description["de-DE"], Markdown: fixture.FrameContent}}
	for _, seed := range seeds {
		if err := engine.writeFrame(seed); err != nil {
			return err
		}
	}

	return nil
}

func (engine *compositionEngine) writeFrame(seed frameSeed) error {
	if err := value.IdentifierSyntax(seed.ID).Validate(); err != nil {
		return err
	}

	metadata := map[string]any{"title": seed.Title, "description": seed.Description, "url": "/_frame/" + seed.Kind + "/" + seed.ID + "/" + strconv.FormatUint(seed.Revision, 10) + "/", "layout": "dynamic", "canonical": seed.Path, "fragment": "/fragments" + seed.Path, "frame_kind": seed.Kind, "root_version": seed.Revision}
	frontmatter, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// #nosec G703 -- Validierte ID und eigenes temporäres Fixtureverzeichnis.
	return os.WriteFile(engine.source+"/content/"+seed.ID+".md", []byte(fmt.Sprintf("%s\n\n%s\n", frontmatter, seed.Markdown)), 0600)
}
