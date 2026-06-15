package cellphones

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "cellphones" {
		t.Errorf("Scheme = %q, want cellphones", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "cellphones" {
		t.Errorf("Identity.Binary = %q, want cellphones", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		{"iphone-16-pro-max-12345", "product", "iphone-16-pro-max-12345"},
		{"iphone-16-pro-max-12345.html", "product", "iphone-16-pro-max-12345"},
		{"https://" + Host + "/samsung-galaxy-s25-67890.html", "product", "samsung-galaxy-s25-67890"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("product", "iphone-16-pro-max-12345")
	want := baseURL + "/iphone-16-pro-max-12345.html"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("unknown", "x")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}

func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}

	p := &Product{
		Slug:  "iphone-16-pro-max-12345",
		URL:   baseURL + "/iphone-16-pro-max-12345.html",
		Name:  "Apple iPhone 16 Pro Max",
		Price: 32990000,
	}
	u, err := h.Mint(p)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	want := "cellphones://product/iphone-16-pro-max-12345"
	if u.String() != want {
		t.Errorf("Mint = %q, want %q", u.String(), want)
	}

	got, err := h.ResolveOn("cellphones", "samsung-galaxy-s25-67890")
	if err != nil || got.String() != "cellphones://product/samsung-galaxy-s25-67890" {
		t.Errorf("ResolveOn = (%q, %v), want cellphones://product/samsung-galaxy-s25-67890", got.String(), err)
	}
}
