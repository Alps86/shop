package opensearch

import (
	"context"
	"encoding/json"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	Endpoint      string
	Index         string
	Context       value.RequestContext
	Schema        application.SearchSchema
	HTTP          *http.Client
	Clock         application.Clock
	PurchaseCosts map[value.SKU]value.Money
	Categories    map[value.CategoryID]catalog.Category
	Source        ReindexSource
}

type ReindexSource interface {
	Snapshot(context.Context, time.Time) (event.ReindexSnapshot, error)
	Catchup(context.Context, string, application.ProjectionPort) error
	Abort(context.Context, string) error
}

type Adapter struct {
	config     Config
	projection *ProjectionState
}

type ProjectionState struct {
	mutex sync.Mutex
	index string
	err   error
}

type Clock struct{}

type JSON map[string]any

type Reply struct {
	Status int
	Body   []byte
}

type SearchResponse struct {
	Hits         HitPage                      `json:"hits"`
	Aggregations map[string]json.RawMessage   `json:"aggregations"`
	Suggest      map[string][]SuggestionGroup `json:"suggest"`
}

type HitPage struct {
	Total HitTotal `json:"total"`
	Hits  []Hit    `json:"hits"`
}

type HitTotal struct {
	Value int64 `json:"value"`
}

type Hit struct {
	Source Document             `json:"_source"`
	Inner  map[string]InnerPage `json:"inner_hits"`
}

type InnerPage struct {
	Hits VariantPage `json:"hits"`
}

type VariantPage struct {
	Hits []VariantHit `json:"hits"`
}

type VariantHit struct {
	Source Variant `json:"_source"`
}

type SuggestionGroup struct {
	Options []SuggestionOption `json:"options"`
}

type SuggestionOption struct {
	Text string `json:"text"`
}

type SuggestionSet struct {
	Values []application.Suggestion
	Seen   map[string]bool
}

type Document struct {
	ID          value.ProductID    `json:"product_id"`
	Revision    uint64             `json:"revision"`
	Deleted     bool               `json:"deleted"`
	Published   bool               `json:"published"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Brand       string             `json:"brand"`
	CategoryIDs []value.CategoryID `json:"category_ids"`
	Variants    []Variant          `json:"variants"`
	Suggest     []string           `json:"suggest"`
	Visibility  string             `json:"visibility"`
}

type Variant struct {
	Margin    int64                         `json:"margin"`
	SKU       value.SKU                     `json:"sku"`
	Price     int64                         `json:"price"`
	Available int64                         `json:"available"`
	Options   map[string][]value.OptionCode `json:"options"`
	Numbers   map[string]int64              `json:"numbers"`
	Card      application.ProductCard       `json:"card"`
}

type FailureResponse struct {
	Error FailureDetail `json:"error"`
}

type FailureDetail struct {
	Type string `json:"type"`
}

type FilterPlan struct {
	Query      application.SearchQuery
	Omit       string
	Predicates map[string]JSON
}

type FacetResponse struct {
	Values   BucketGroup   `json:"values"`
	Variants FacetVariants `json:"variants"`
}

type FacetVariants struct {
	Matching FacetMatches `json:"matching"`
}

type FacetMatches struct {
	Values BucketGroup `json:"values"`
}

type BucketGroup struct {
	Buckets []FacetBucket `json:"buckets"`
}

type FacetBucket struct {
	Key      json.RawMessage `json:"key"`
	Count    int64           `json:"doc_count"`
	Products ProductCount    `json:"products"`
}

type ProductCount struct {
	Count int64 `json:"doc_count"`
}

type Ranking struct {
	Interests []sale.CategoryInterest
}

type multiSearchLines []JSON

type searchRequest struct {
	Path    string
	Payload any
	Multi   bool
}

type multiSearchResponse struct {
	Responses []multiSearchPart `json:"responses"`
}

type multiSearchPart struct {
	SearchResponse
	Status   int             `json:"status"`
	Error    json.RawMessage `json:"error"`
	TimedOut bool            `json:"timed_out"`
	Shards   searchShards    `json:"_shards"`
}

type searchShards struct {
	Failed int `json:"failed"`
}
