package handler

import (
	"testing"
	"time"
)

func TestImportStatus(t *testing.T) {
	if got := importStatus(nil); got != "processing" {
		t.Fatalf("status sem sincronização = %q", got)
	}
	now := time.Now()
	if got := importStatus(&now); got != "completed" {
		t.Fatalf("status sincronizado = %q", got)
	}
}
