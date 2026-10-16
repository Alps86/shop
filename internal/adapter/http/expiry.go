package http

import (
	"fmt"
	"github.com/Alps86/shop/internal/application"
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

	window := cacheWindow{now}
	ttl := min(window.cardsTTL(page.Search.Cards), window.cardsTTL(page.Detail.Cards), window.cardsTTL(page.Recommendations.Cards))

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

func (window cacheWindow) cardsTTL(cards []application.ProductCard) int {
	ttl := 5
	for _, card := range cards {
		ttl = min(ttl, window.complianceTTL(card.Compliance))
	}

	return ttl
}
