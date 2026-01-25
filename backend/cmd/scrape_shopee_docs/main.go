package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type moduleListResponse struct {
	Modules []docModule `json:"modules"`
}

type docModule struct {
	ModuleID   int       `json:"module_id"`
	ModuleName string    `json:"module_name"`
	Items      []docItem `json:"items"`
}

type docItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"` // 1 = API doc, 2 = guide/article
}

type apiDoc struct {
	APIID      int    `json:"api_id"`
	APIName    string `json:"api_name"`
	ModuleID   int    `json:"module_id"`
	ModuleName string `json:"module_name"`
	Method     int    `json:"method"`
	Path       string `json:"path"`
	URL        string `json:"url"`
	TestURL    string `json:"test_url"`

	Define         string `json:"define"`
	Params         string `json:"params"`          // JSON string with request/response params.
	RequestSample  string `json:"request_sample"`  // JSON string.
	ResponseSample string `json:"response_sample"` // JSON string.
	ErrorExample   string `json:"error_example"`   // JSON string.
}

type output struct {
	GeneratedAt time.Time     `json:"generated_at"`
	Version     int           `json:"version"`
	Modules     []moduleIndex `json:"modules"`
	APIs        []apiDoc      `json:"apis"`
}

type moduleIndex struct {
	ModuleID   int      `json:"module_id"`
	ModuleName string   `json:"module_name"`
	APIIDs     []int    `json:"api_ids"`
	APINames   []string `json:"api_names"`
}

func main() {
	var (
		baseURL = flag.String("base", "https://open.shopee.com/api/v1", "Shopee docs base URL")
		version = flag.Int("version", 2, "Docs version (1 or 2)")
		outPath = flag.String("out", filepath.FromSlash("docs/shopee_openplatform_v2.json"), "Output JSON path (relative to backend/)")
		sleepMs = flag.Int("sleep-ms", 100, "Sleep between API doc fetches (ms)")
		maxAPIs = flag.Int("max", 0, "Max number of APIs to fetch (0 = no limit)")
	)
	flag.Parse()

	client := &http.Client{Timeout: 30 * time.Second}

	modulesURL := fmt.Sprintf("%s/doc/module/?version=%d", *baseURL, *version)
	var moduleResp moduleListResponse
	if err := getJSON(client, modulesURL, &moduleResp); err != nil {
		fatalf("fetch module list: %v", err)
	}

	var (
		out     output
		apiDocs []apiDoc
	)
	out.GeneratedAt = time.Now().UTC()
	out.Version = *version

	apiCount := 0
	for _, m := range moduleResp.Modules {
		mi := moduleIndex{ModuleID: m.ModuleID, ModuleName: m.ModuleName}

		for _, it := range m.Items {
			if it.Type != 1 {
				continue
			}
			if *maxAPIs > 0 && apiCount >= *maxAPIs {
				break
			}

			docURL := fmt.Sprintf("%s/doc/api/?api_id=%d", *baseURL, it.ID)
			var doc apiDoc
			if err := getJSON(client, docURL, &doc); err != nil {
				// Best-effort: continue so one bad doc doesn't stop the whole scrape.
				fmt.Fprintf(os.Stderr, "[WARN] fetch api_id=%d (%s): %v\n", it.ID, it.Name, err)
				continue
			}

			apiDocs = append(apiDocs, doc)
			mi.APIIDs = append(mi.APIIDs, it.ID)
			mi.APINames = append(mi.APINames, it.Name)

			apiCount++
			if *sleepMs > 0 {
				time.Sleep(time.Duration(*sleepMs) * time.Millisecond)
			}
		}

		out.Modules = append(out.Modules, mi)
		if *maxAPIs > 0 && apiCount >= *maxAPIs {
			break
		}
	}
	out.APIs = apiDocs

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fatalf("mkdir: %v", err)
	}
	f, err := os.Create(*outPath)
	if err != nil {
		fatalf("create output: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fatalf("write output: %v", err)
	}

	fmt.Printf("Wrote %d APIs to %s\n", len(out.APIs), *outPath)
}

func getJSON(client *http.Client, url string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "omni-backend-scraper/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return fmt.Errorf("status=%d body=%q", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
