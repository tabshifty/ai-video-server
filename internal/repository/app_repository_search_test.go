package repository

import (
	"strings"
	"testing"
)

func TestSearchVideosCountSQLAvoidsDistinct(t *testing.T) {
	t.Parallel()
	if strings.Contains(strings.ToUpper(searchVideosCountSQL), "DISTINCT") {
		t.Fatalf("count SQL must not use DISTINCT, got: %s", searchVideosCountSQL)
	}
}

func TestSearchVideosCountSQLUsesExistsForTagMatch(t *testing.T) {
	t.Parallel()
	if !strings.Contains(searchVideosCountSQL, "EXISTS (SELECT 1 FROM video_tags") {
		t.Fatalf("count SQL must use EXISTS for tag matching, got: %s", searchVideosCountSQL)
	}
}

func TestSearchVideosSelectSQLAvoidsDistinctForReleaseDateSort(t *testing.T) {
	t.Parallel()
	releaseOrder := "NULLIF(v.metadata->>'release_date', '')::date DESC NULLS LAST, v.created_at DESC"
	sql := searchVideosSelectSQL(releaseOrder)
	if strings.Contains(strings.ToUpper(sql), "SELECT DISTINCT") {
		t.Fatalf("select SQL must not use SELECT DISTINCT (triggers Postgres 42P10 when order expression is not in select list), got: %s", sql)
	}
}

func TestSearchVideosSelectSQLEmbedsOrderClauseVerbatim(t *testing.T) {
	t.Parallel()
	releaseOrder := "NULLIF(v.metadata->>'release_date', '')::date DESC NULLS LAST, v.created_at DESC"
	sql := searchVideosSelectSQL(releaseOrder)
	if !strings.Contains(sql, "ORDER BY "+releaseOrder) {
		t.Fatalf("select SQL must embed order clause verbatim, got: %s", sql)
	}
}

func TestSearchVideosSelectSQLUsesExistsForTagMatch(t *testing.T) {
	t.Parallel()
	sql := searchVideosSelectSQL("v.created_at DESC")
	if !strings.Contains(sql, "EXISTS (SELECT 1 FROM video_tags") {
		t.Fatalf("select SQL must use EXISTS for tag matching, got: %s", sql)
	}
}
