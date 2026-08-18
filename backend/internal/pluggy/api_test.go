package pluggy

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestResponseErrorIncludesPluggyReason(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusConflict,
		Body: io.NopCloser(strings.NewReader(
			`{"codeDescription":"BEFORE_ALLOWED_FREQUENCY","message":"wait one hour"}`,
		)),
	}
	got := responseError("update", resp).Error()
	if !strings.Contains(got, "409") || !strings.Contains(got, "BEFORE_ALLOWED_FREQUENCY") {
		t.Fatalf("erro sem motivo da Pluggy: %q", got)
	}
}
