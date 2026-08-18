package tests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wyw/cry031-volunteer/internal/application"
	"github.com/wyw/cry031-volunteer/internal/repository"
	"github.com/wyw/cry031-volunteer/internal/service"
	httptransport "github.com/wyw/cry031-volunteer/internal/transport/http"
	"go.uber.org/zap"
)

func testRouter(t *testing.T) http.Handler {
	now := time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC)
	store := repository.NewMemoryStore(repository.DemoState(now))
	engine := application.NewEngine(store, service.FixedClock{Time: now}, &service.LocalNotifier{})
	attachments := service.LocalFileStore{Root: t.TempDir(), MaxBytes: 5 << 20}
	return httptransport.NewRouter(engine, attachments, zap.NewNop(), 5*time.Second, func() bool { return true })
}

func TestHealthAndDiscoveryEndpoints(t *testing.T) {
	router := testRouter(t)
	for _, path := range []string{"/healthz", "/readyz", "/api/v1/catalog", "/api/v1/discovery/teams?page=1&page_size=10"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, body=%s", path, res.Code, res.Body.String())
		}
		if path == "/api/v1/discovery/teams?page=1&page_size=10" && strings.Contains(res.Body.String(), "RIVER-2026") {
			t.Fatal("discovery endpoint leaked invite code")
		}
	}
}

func TestActivityCreationRequiresManager(t *testing.T) {
	router := testRouter(t)
	body := `{"team_id":"team-riverside","title":"周末服务","description":"test","start_at":"2026-08-20T09:00:00Z","end_at":"2026-08-20T10:00:00Z","location":"服务站"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/activities", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "u-volunteer")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("member create status = %d, body=%s", res.Code, res.Body.String())
	}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/activities", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "u-captain")
	req.Header.Set("X-User-Role", "captain")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("captain create status = %d, body=%s", res.Code, res.Body.String())
	}
}

func TestValidationErrorIncludesRequestIDAndFields(t *testing.T) {
	router := testRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/activities", strings.NewReader(`{"team_id":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "request-test-1")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("validation status = %d", res.Code)
	}
	if !strings.Contains(res.Body.String(), "request-test-1") || !strings.Contains(res.Body.String(), "TeamID") {
		t.Fatalf("validation response lacks request metadata: %s", res.Body.String())
	}
}

func TestUnknownSortReturnsStableClientError(t *testing.T) {
	router := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/discovery/teams?sort=unsafe", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest || !strings.Contains(res.Body.String(), "invalid_filter") {
		t.Fatalf("unexpected sort response %d: %s", res.Code, res.Body.String())
	}
}

func TestRoleHeaderCannotEscalateProfileAccess(t *testing.T) {
	router := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/u-captain/records", nil)
	req.Header.Set("X-User-ID", "u-volunteer")
	req.Header.Set("X-User-Role", "admin")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("spoofed admin status = %d, body=%s", res.Code, res.Body.String())
	}
}

func TestAttachmentUploadReturnsOpaqueLocalToken(t *testing.T) {
	router := testRouter(t)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "现场记录.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("png-data"))
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", "u-volunteer")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusCreated || !strings.Contains(res.Body.String(), "attachment_token") {
		t.Fatalf("upload status = %d, body=%s", res.Code, res.Body.String())
	}
}
