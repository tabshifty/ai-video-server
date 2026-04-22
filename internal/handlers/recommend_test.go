package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestParseUUIDCSV(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	got, err := parseUUIDCSV(id1.String() + "," + id2.String() + "," + id1.String())
	if err != nil {
		t.Fatalf("parseUUIDCSV returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 unique ids, got %d", len(got))
	}
	if got[0] != id1 || got[1] != id2 {
		t.Fatalf("unexpected parse result: %#v", got)
	}
}

func TestParseUUIDCSVRejectsInvalidID(t *testing.T) {
	if _, err := parseUUIDCSV("not-a-uuid"); err == nil {
		t.Fatal("expected invalid uuid error")
	}
}

func TestRegisterIncludesImageCollectionRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := &API{}
	api.Register(router)

	routes := map[string]struct{}{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	for _, want := range []string{
		"GET /api/v1/image-collections",
		"GET /api/v1/image-collections/:id",
		"GET /api/v1/images/:id/view",
	} {
		if _, ok := routes[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}
