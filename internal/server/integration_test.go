package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// These tests run against a real PostgreSQL database.
// Set TEST_DATABASE_URL to a database that can be wiped, e.g.
// postgres://suiten:suiten_test@localhost:5432/suiten_test?sslmode=disable

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	database, err := sqlx.Connect("postgres", url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if _, err := database.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	applyMigrations(t, database, "up")
	return database
}

func migrationFiles(t *testing.T, direction string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*."+direction+".sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no %s migrations found: %v", direction, err)
	}
	sort.Strings(files)
	if direction == "down" {
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
	}
	return files
}

func applyMigrations(t *testing.T, database *sqlx.DB, direction string) {
	t.Helper()
	for _, f := range migrationFiles(t, direction) {
		sql, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if _, err := database.Exec(string(sql)); err != nil {
			t.Fatalf("apply %s: %v", filepath.Base(f), err)
		}
	}
}

type apiClient struct {
	t     *testing.T
	srv   http.Handler
	token string
}

func (c *apiClient) do(method, path string, body interface{}) map[string]interface{} {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			c.t.Fatalf("encode: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	rec := httptest.NewRecorder()
	c.srv.ServeHTTP(rec, req)
	if rec.Code >= 300 {
		c.t.Fatalf("%s %s: status %d: %s", method, path, rec.Code, rec.Body.String())
	}
	var out map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		c.t.Fatalf("decode %s: %v", rec.Body.String(), err)
	}
	return out
}

func login(t *testing.T, c *apiClient) {
	t.Helper()
	c.do(http.MethodPost, "/v1/auth/register", map[string]string{
		"email": "reader@example.com", "password": "correct-horse-battery",
	})
	res := c.do(http.MethodPost, "/v1/auth/login", map[string]string{
		"email": "reader@example.com", "password": "correct-horse-battery",
		"device_id": uuid.NewString(), "device_name": "test",
	})
	data := res["data"].(map[string]interface{})
	c.token = data["access_token"].(string)
}

func syncBody(bookID, sessionID uuid.UUID, captures []uuid.UUID, emotions []string, book map[string]interface{}) map[string]interface{} {
	capItems := make([]map[string]interface{}, 0, len(captures))
	now := time.Now()
	for i, id := range captures {
		capItems = append(capItems, map[string]interface{}{
			"id": id, "action": "upsert",
			"data": map[string]interface{}{
				"session_id": sessionID, "emotion": emotions[i], "user_memo": "メモ",
				"ai_status": "done", "created_at": now.Unix(), "updated_at": now.Unix(),
			},
		})
	}
	return map[string]interface{}{
		"last_sync_at": 0,
		"device_id":    "test",
		"changes": map[string]interface{}{
			"books": []map[string]interface{}{{"id": bookID, "action": "upsert", "data": book}},
			"reading_sessions": []map[string]interface{}{{
				"id": sessionID, "action": "upsert",
				"data": map[string]interface{}{
					"book_id": bookID, "started_at": now.Unix(), "created_at": now.Unix(), "updated_at": now.Unix(),
				},
			}},
			"captures": capItems,
		},
	}
}

func findBook(t *testing.T, res map[string]interface{}, id uuid.UUID) map[string]interface{} {
	t.Helper()
	changes := res["data"].(map[string]interface{})["changes"].(map[string]interface{})
	for _, b := range changes["books"].([]interface{}) {
		book := b.(map[string]interface{})
		if book["id"] == id.String() {
			return book
		}
	}
	t.Fatalf("book %s not in sync response", id)
	return nil
}

func TestSyncRecordsSummaryCompletionAndNorthStar(t *testing.T) {
	database := openTestDB(t)
	c := &apiClient{t: t, srv: New(database)}
	login(t, c)

	now := time.Now()
	bookID, sessionID := uuid.New(), uuid.New()
	captures := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	emotions := []string{"🤔", "❤️", "💡,🤔,⚠️,🔥,❤️"}

	// Book timestamps in milliseconds, as older clients sent them.
	book := map[string]interface{}{
		"title":                 "小さく始める新規事業",
		"summary":               "疑問は意思決定の仕組みに、共感は手応えに向いていたようです。",
		"created_at":            now.UnixMilli(),
		"updated_at":            now.UnixMilli(),
		"summary_generated_at":  now.Unix(),
		"summary_saved_at":      now.Unix(),
		"summary_helpfulness":   3,
		"summary_capture_count": 3,
	}
	res := c.do(http.MethodPost, "/v1/sync", syncBody(bookID, sessionID, captures, emotions, book))

	got := findBook(t, res, bookID)
	if got["created_at"].(float64) != float64(now.Unix()) {
		t.Errorf("created_at = %v, want %d (milliseconds should be read as the same instant)", got["created_at"], now.Unix())
	}
	completedAt, ok := got["summary_completed_at"].(float64)
	if !ok {
		t.Fatalf("summary_completed_at missing in %v", got)
	}

	// Lowering the rating later keeps the first completion time.
	book["summary_helpfulness"] = 1
	book["updated_at"] = now.Add(time.Minute).Unix()
	res = c.do(http.MethodPost, "/v1/sync", syncBody(bookID, sessionID, captures, emotions, book))
	if got := findBook(t, res, bookID); got["summary_completed_at"].(float64) != completedAt {
		t.Errorf("summary_completed_at changed from %v to %v", completedAt, got["summary_completed_at"])
	}

	analytics := c.do(http.MethodGet, "/v1/analytics/personal", nil)
	ns := analytics["data"].(map[string]interface{})["north_star"].(map[string]interface{})
	monthly := ns["monthly"].([]interface{})
	current := monthly[len(monthly)-1].(map[string]interface{})
	if current["month"] != now.UTC().Format("2006-01") || current["completed_summaries"].(float64) != 1 {
		t.Errorf("current month = %v, want %s with 1 completed summary", current, now.UTC().Format("2006-01"))
	}

	inputs := ns["inputs"].(map[string]interface{})
	for key, want := range map[string]float64{
		"active_books":           1,
		"captures_per_book":      3,
		"doubt_and_empathy_rate": 1,
		"summary_creation_rate":  1,
	} {
		if inputs[key].(float64) != want {
			t.Errorf("inputs[%s] = %v, want %v", key, inputs[key], want)
		}
	}

	global := c.do(http.MethodGet, "/v1/analytics/global", nil)
	if _, ok := global["data"].(map[string]interface{})["north_star"].(map[string]interface{}); !ok {
		t.Error("global analytics should include north_star")
	}
}

func TestIncompleteSummaryIsNotCounted(t *testing.T) {
	database := openTestDB(t)
	c := &apiClient{t: t, srv: New(database)}
	login(t, c)

	now := time.Now()
	bookID, sessionID := uuid.New(), uuid.New()
	captures := []uuid.UUID{uuid.New(), uuid.New()}
	book := map[string]interface{}{
		"title": "二件だけの本", "created_at": now.Unix(), "updated_at": now.Unix(),
		"summary_generated_at": now.Unix(), "summary_saved_at": now.Unix(),
		"summary_helpfulness": 3, "summary_capture_count": 2,
	}
	res := c.do(http.MethodPost, "/v1/sync", syncBody(bookID, sessionID, captures, []string{"🤔", "💡"}, book))
	if v, ok := findBook(t, res, bookID)["summary_completed_at"]; ok {
		t.Errorf("summary with 2 captures should not be complete, got summary_completed_at=%v", v)
	}
}

func TestMigrationsRoundTrip(t *testing.T) {
	database := openTestDB(t)
	applyMigrations(t, database, "down")
	var tables int
	if err := database.Get(&tables, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'`); err != nil {
		t.Fatal(err)
	}
	if tables != 0 {
		t.Errorf("%d tables left after all down migrations", tables)
	}
	applyMigrations(t, database, "up")
}

func TestHealth(t *testing.T) {
	database := openTestDB(t)
	rec := httptest.NewRecorder()
	New(database).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "healthy") {
		t.Errorf("GET /health = %d %s", rec.Code, rec.Body.String())
	}
}
