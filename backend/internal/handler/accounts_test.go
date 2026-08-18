package handler

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/finance-os/backend/internal/service"
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

func TestSyncRunStatus(t *testing.T) {
	for name, test := range map[string]struct {
		err  error
		want string
	}{
		"completed": {want: "completed"},
		"partial":   {err: fmt.Errorf("detalhe: %w", service.ErrPartialSync), want: "partial"},
		"failed":    {err: errors.New("falha"), want: "failed"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := syncRunStatus(test.err); got != test.want {
				t.Fatalf("syncRunStatus() = %q, quer %q", got, test.want)
			}
		})
	}
}
