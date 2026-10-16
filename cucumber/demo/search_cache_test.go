package demo_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (current *world) registerSearchCache(ctx *godog.ScenarioContext) {
	ctx.Step(`^die öffentlichen Suchantworten sind tatsächlich gespeichert$`, current.storeSearchCache)
	ctx.Step(`^ein neues Produkt wird über den echten Importport veröffentlicht$`, current.importSearchProduct)
	ctx.Step(`^das hinzugefügte Produkt wird über den echten Katalogport gelöscht$`, current.deleteSearchProduct)
	ctx.Step(`^die bestätigte Projektion invalidiert die gespeicherten Suchen vor Ablauf$`, current.observeSearchInvalidation)
	ctx.Step(`^die bisherigen Lagerwerte bleiben unverändert$`, current.observeSearchInventory)
	ctx.Step(`^der Lagerdatensatz ist nach der Produktlöschung weiterhin vorhanden$`, current.retainedSearchInventory)
}

func (current *world) searchGet(path string) error {
	request, err := http.NewRequest(http.MethodGet, current.deployment.public+path, nil)
	if err != nil {
		return err
	}

	request.Host = current.searchCache.host

	return current.send(request)
}

func (current *world) storeSearchCache(table *godog.Table) error {
	data := read(table)[0]
	current.searchCache.host = data["HostPrefix"] + "-" + strings.ToLower(rand.Text()) + ".local"

	if err := current.primeSearchEntries(table); err != nil {
		return err
	}

	bans, err := current.deployment.cacheCounter("MAIN.bans_added")
	current.searchCache.bans = bans

	return err
}

func (current *world) primeSearchEntries(table *godog.Table) error {
	before, err := current.deployment.cacheHits()
	if err != nil {
		return err
	}

	current.searchCache.entries = nil
	for _, row := range read(table) {
		if err := current.storeSearchEntry(row); err != nil {
			return err
		}
	}

	return current.awaitHits(before, read(table)[0])
}

func (current *world) storeSearchEntry(data map[string]string) error {
	ttl, err := strconv.Atoi(data["TTLSeconds"])
	if err != nil {
		return err
	}

	entry := searchCacheEntry{path: data["Path"], stored: time.Now(), ttl: time.Duration(ttl) * time.Second}
	if err := current.searchGet(entry.path); err != nil {
		return err
	}

	if err := current.observeStoredSearch(data); err != nil {
		return err
	}

	current.searchCache.entries = append(current.searchCache.entries, entry)

	return current.searchGet(entry.path)
}

func (current *world) observeStoredSearch(data map[string]string) error {
	if err := current.observeRow(data); err != nil {
		return err
	}

	keys := strings.Fields(current.response.header.Get("Surrogate-Key"))
	if err := searchTagCount(keys, data["TagCount"]); err != nil {
		return err
	}

	for _, key := range keys {
		if key == data["Tag"] {
			return nil
		}
	}

	return fmt.Errorf("Suchcachetag %q fehlt in %v", data["Tag"], keys)
}

func searchTagCount(keys []string, expected string) error {
	if expected == "" {
		return nil
	}

	count, err := strconv.Atoi(expected)
	if err != nil {
		return err
	}

	if len(keys) != count {
		return fmt.Errorf("Suchantwort hat %d statt %d Cachetags", len(keys), count)
	}

	return nil
}

func (current *world) observeSearchInvalidation(table *godog.Table) error {
	data := read(table)[0]
	seconds, err := strconv.Atoi(data["TimeoutSeconds"])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	defer cancel()
	if err := current.waitSearchProjection(ctx, data); err != nil {
		return err
	}

	return current.observeFreshSearchRows(table)
}

func (current *world) observeFreshSearchRows(table *godog.Table) error {
	for _, row := range read(table) {
		if err := current.observeFreshSearch(row); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) waitSearchProjection(ctx context.Context, data map[string]string) error {
	minimum, err := strconv.ParseInt(data["MinimumBans"], 10, 64)
	if err != nil {
		return err
	}

	for ctx.Err() == nil {
		ready, err := current.searchProjectionBanned(ctx, data, minimum)
		if err != nil {
			return err
		}

		if ready {
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("echte Projektion und bestätigter Ban fehlen: %w", ctx.Err())
}

func (current *world) searchProjectionBanned(ctx context.Context, data map[string]string, minimum int64) (bool, error) {
	if !current.searchProjected(ctx, data) {
		return false, nil
	}

	bans, err := current.deployment.cacheCounter("MAIN.bans_added")

	return bans-current.searchCache.bans >= minimum, err
}

func (current *world) searchProjected(ctx context.Context, data map[string]string) bool {
	reply, err := current.readSearchProjection(ctx, data["ProductID"])
	if err != nil {
		return false
	}

	if !reply.matches(data) {
		return false
	}

	checkpoint, err := current.deployment.store.Checkpoint(ctx, current.searchCache.slot)
	if err != nil {
		return false
	}

	return checkpoint > current.searchCache.checkpoint
}

func (reply searchProjectionReply) matches(data map[string]string) bool {
	if strconv.FormatBool(reply.Source.Deleted) != data["Deleted"] {
		return false
	}

	return reply.Source.ID == value.ProductID(data["ProductID"])
}

func (current *world) readSearchProjection(ctx context.Context, id string) (searchProjectionReply, error) {
	path := current.deployment.config.SearchEndpoint + "/" + url.PathEscape(current.deployment.config.SearchIndex) + "/_doc/" + url.PathEscape(id)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return searchProjectionReply{}, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return searchProjectionReply{}, err
	}

	defer func() {
		_ = response.Body.Close()
	}()
	var reply searchProjectionReply
	err = json.NewDecoder(response.Body).Decode(&reply)

	return reply, err
}

func (current *world) observeFreshSearch(data map[string]string) error {
	for _, entry := range current.searchCache.entries {
		if entry.path == data["Path"] {
			return entry.observe(current, data)
		}
	}

	return fmt.Errorf("keine gespeicherte Suchantwort für %q", data["Path"])
}

func (entry searchCacheEntry) observe(current *world, data map[string]string) error {
	if err := entry.fresh(); err != nil {
		return err
	}

	if err := current.searchGet(entry.path); err != nil {
		return err
	}

	if err := entry.fresh(); err != nil {
		return err
	}

	return current.observeRow(data)
}

func (entry searchCacheEntry) fresh() error {
	if time.Since(entry.stored) >= entry.ttl {
		return fmt.Errorf("Suchcache lief vor abgeschlossenem Invalidierungsnachweis ab")
	}

	return nil
}
