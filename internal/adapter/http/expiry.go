package http

import (
	"fmt"
	"github.com/Alps86/shop/internal/domain/catalog"
	"strings"
	"time"
)

func (clock systemClock) Now() time.Time {
	return time.Now()
}

func (page *Page) limitCache(now time.Time) {
	if !strings.HasPrefix(page.CacheControl, "public,max-age=5") {
		return
	}

	ttl := 5
	for _, card := range page.Search.Cards {
		ttl = min(ttl, (cacheWindow{now}).complianceTTL(card.Compliance))
	}

	for _, card := range page.Detail.Cards {
		ttl = min(ttl, (cacheWindow{now}).complianceTTL(card.Compliance))
	}

	page.CacheControl = fmt.Sprintf("public,max-age=%d", ttl)
	if ttl <= 0 {
		page.CacheControl = "private,no-store"
	}
}

func (window cacheWindow) complianceTTL(compliance catalog.ProductCompliance) int {
	now := window.now
	ttl := 5
	if compliance.Context.ValidUntil.Present {
		ttl = min(ttl, int(compliance.Context.ValidUntil.Value.Sub(now)/time.Second))
	}

	for _, rule := range compliance.ApplicableRules {
		ttl = min(ttl, window.ruleTTL(rule))
	}

	return max(0, ttl)
}

func (window cacheWindow) ruleTTL(rule catalog.RuleReference) int {
	now := window.now
	ttl := 5
	if rule.ValidUntil.Present {
		ttl = min(ttl, int(rule.ValidUntil.Value.Sub(now)/time.Second))
	}

	if rule.ValidFrom.Present {
		if rule.ValidFrom.Value.After(now) {
			ttl = min(ttl, int(rule.ValidFrom.Value.Sub(now)/time.Second))
		}
	}

	return max(0, ttl)
}
