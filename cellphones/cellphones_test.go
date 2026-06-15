package cellphones

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(srv *httptest.Server) *Client {
	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 0
	return NewClientWithConfig(cfg)
}

func sampleProductPageHTML(slug, name, brand string, warranty int, genuine bool) string {
	ld, _ := json.Marshal(map[string]any{
		"@type":       "Product",
		"name":        name,
		"description": "Great phone",
		"brand":       map[string]string{"name": brand},
		"offers":      map[string]string{"price": "29990000", "highPrice": "32990000"},
		"aggregateRating": map[string]string{
			"ratingValue": "4.9",
			"reviewCount": "2345",
		},
	})
	genuineText := ""
	if genuine {
		genuineText = `<span class="label">Hàng Chính Hãng</span>`
	}
	warrantyText := ""
	if warranty > 0 {
		warrantyText = `<div class="warranty">Bảo hành 12 tháng</div>`
	}
	return `<!DOCTYPE html><html><head>
<script type="application/ld+json">` + string(ld) + `</script>
</head><body>` + genuineText + warrantyText + `</body></html>`
}

func sampleProductListJSON(n int, categoryID string) string {
	type item struct {
		ID          int64   `json:"id"`
		Name        string  `json:"name"`
		Slug        string  `json:"slug"`
		Price       float64 `json:"price"`
		OldPrice    float64 `json:"old_price"`
		Brand       string  `json:"brand"`
		Category    string  `json:"category"`
		RatingPoint float64 `json:"rating_point"`
		ReviewCount int     `json:"review_count"`
		IsGenuine   bool    `json:"is_genuine"`
	}
	var items []item
	for i := 0; i < n; i++ {
		items = append(items, item{
			ID:          int64(10000 + i),
			Name:        "Phone " + string(rune('A'+i)),
			Slug:        "phone-" + string(rune('a'+i)) + "-12345",
			Price:       float64(10000000 * (i + 1)),
			OldPrice:    float64(12000000 * (i + 1)),
			Brand:       "Samsung",
			Category:    "Điện thoại",
			RatingPoint: 4.5,
			ReviewCount: 100,
			IsGenuine:   true,
		})
	}
	b, _ := json.Marshal(map[string]any{
		"products":    items,
		"total_page":  3,
		"total_count": n,
	})
	return string(b)
}

func sampleReviewListJSON(n int) string {
	type r struct {
		ReviewID    int64  `json:"review_id"`
		FullName    string `json:"full_name"`
		RatingPoint int    `json:"rating_point"`
		Content     string `json:"content"`
	}
	var list []r
	for i := 0; i < n; i++ {
		list = append(list, r{
			ReviewID:    int64(5000 + i),
			FullName:    "User " + string(rune('A'+i)),
			RatingPoint: 5,
			Content:     "Excellent",
		})
	}
	b, _ := json.Marshal(map[string]any{"data": list})
	return string(b)
}

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "ok" {
		t.Errorf("body = %q, want ok", body)
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("recovered"))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := NewClientWithConfig(cfg)

	start := time.Now()
	body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "recovered" {
		t.Errorf("body = %q after retries", body)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestGetProduct(t *testing.T) {
	html := sampleProductPageHTML("samsung-galaxy-s25-12345", "Samsung Galaxy S25", "Samsung", 12, true)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	p, err := c.GetProduct(context.Background(), "samsung-galaxy-s25-12345")
	if err != nil {
		t.Fatal(err)
	}
	if p.Slug != "samsung-galaxy-s25-12345" {
		t.Errorf("Slug = %q", p.Slug)
	}
	if p.Name != "Samsung Galaxy S25" {
		t.Errorf("Name = %q", p.Name)
	}
	if p.Brand != "Samsung" {
		t.Errorf("Brand = %q", p.Brand)
	}
	if p.Rating != 4.9 {
		t.Errorf("Rating = %v, want 4.9", p.Rating)
	}
	if p.ReviewCount != 2345 {
		t.Errorf("ReviewCount = %d, want 2345", p.ReviewCount)
	}
	if !p.IsGenuine {
		t.Error("IsGenuine = false, want true")
	}
	if p.WarrantyMonths != 12 {
		t.Errorf("WarrantyMonths = %d, want 12", p.WarrantyMonths)
	}
}

func TestListProducts(t *testing.T) {
	listJSON := sampleProductListJSON(3, "1")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "category=1") {
			t.Errorf("expected category=1 in query, got %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(listJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	products, err := c.ListProducts(context.Background(), "1", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 3 {
		t.Fatalf("got %d products, want 3", len(products))
	}
	if products[0].Brand != "Samsung" {
		t.Errorf("Brand = %q, want Samsung", products[0].Brand)
	}
}

func TestListProductsLimit(t *testing.T) {
	listJSON := sampleProductListJSON(8, "2")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(listJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	products, err := c.ListProducts(context.Background(), "2", 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 4 {
		t.Errorf("got %d products, want 4 (limit)", len(products))
	}
}

func TestListReviews(t *testing.T) {
	reviews := sampleReviewListJSON(3)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(reviews))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	got, err := c.ListReviews(context.Background(), "samsung-galaxy-s25-12345", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d reviews, want 3", len(got))
	}
	if got[0].ID != "5000" {
		t.Errorf("got[0].ID = %q, want 5000", got[0].ID)
	}
}

func TestExtractSlug(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://cellphones.com.vn/iphone-16-pro-max-12345.html", "iphone-16-pro-max-12345"},
		{"iphone-16-pro-max-12345.html", "iphone-16-pro-max-12345"},
		{"iphone-16-pro-max-12345", "iphone-16-pro-max-12345"},
		{"", ""},
	}
	for _, tc := range cases {
		got := extractSlug(tc.in)
		if got != tc.want {
			t.Errorf("extractSlug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
