package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStudioServerEndpoints(t *testing.T) {
	mux := buildStudioMux()

	// 1. Test GET /studio interactive workbench UI
	reqUI := httptest.NewRequest(http.MethodGet, "/studio", nil)
	recUI := httptest.NewRecorder()
	mux.ServeHTTP(recUI, reqUI)
	if recUI.Code != http.StatusOK {
		t.Fatalf("GET /studio status = %d", recUI.Code)
	}
	uiBody := recUI.Body.String()
	for _, token := range []string{"artSelect", "paletteSelect", "paramSlidersContainer", "renderCanvas", "randomizeCurrentArt", "downloadCurrentPNG"} {
		if !strings.Contains(uiBody, token) {
			t.Fatalf("GET /studio missing expected UI element %q", token)
		}
	}

	// 2. Test GET /api/engines
	reqEng := httptest.NewRequest(http.MethodGet, "/api/engines", nil)
	recEng := httptest.NewRecorder()
	mux.ServeHTTP(recEng, reqEng)
	if recEng.Code != http.StatusOK {
		t.Fatalf("GET /api/engines status = %d", recEng.Code)
	}
	var engResp struct {
		Engines       []map[string]interface{} `json:"engines"`
		Palettes      []string                 `json:"palettes"`
		PaletteColors map[string][]string      `json:"palette_colors"`
	}
	if err := json.Unmarshal(recEng.Body.Bytes(), &engResp); err != nil {
		t.Fatalf("unmarshal /api/engines failed: %v", err)
	}
	if len(engResp.Engines) < 140 || len(engResp.Palettes) < 10 || len(engResp.PaletteColors["dunhuang"]) == 0 {
		t.Fatalf("unexpected engine/palette count: engines=%d palettes=%d", len(engResp.Engines), len(engResp.Palettes))
	}

	// 3. Test GET /api/render with random=1 and aesthetic scoring
	reqRender := httptest.NewRequest(http.MethodGet, "/api/render?art=cloisonneenamel&w=64&h=64&seed=2026&random=1", nil)
	recRender := httptest.NewRecorder()
	mux.ServeHTTP(recRender, reqRender)
	if recRender.Code != http.StatusOK {
		t.Fatalf("GET /api/render status = %d: %s", recRender.Code, recRender.Body.String())
	}
	var renderResp struct {
		Art         string                 `json:"art"`
		ImageBase64 string                 `json:"image_base64"`
		CLICommand  string                 `json:"cli_command"`
		Aesthetic   map[string]interface{} `json:"aesthetic"`
	}
	if err := json.Unmarshal(recRender.Body.Bytes(), &renderResp); err != nil {
		t.Fatalf("unmarshal /api/render failed: %v", err)
	}
	if renderResp.Art != "cloisonneenamel" || len(renderResp.ImageBase64) < 100 || renderResp.CLICommand == "" {
		t.Fatalf("unexpected /api/render response: art=%s cmd=%s", renderResp.Art, renderResp.CLICommand)
	}

	// 4. Test GET /api/render with download=1 direct PNG attachment
	reqDL := httptest.NewRequest(http.MethodGet, "/api/render?art=cloisonneenamel&w=64&h=64&seed=2026&download=1", nil)
	recDL := httptest.NewRecorder()
	mux.ServeHTTP(recDL, reqDL)
	if recDL.Code != http.StatusOK || recDL.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("GET /api/render?download=1 status=%d content-type=%s", recDL.Code, recDL.Header().Get("Content-Type"))
	}
	if !strings.Contains(recDL.Header().Get("Content-Disposition"), "attachment;") {
		t.Fatalf("expected Content-Disposition attachment header, got %q", recDL.Header().Get("Content-Disposition"))
	}
}
