package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	Method     int    `json:"method"` // 1 = POST, 2 = GET (per current docs)
	Path       string `json:"path"`
	Params     string `json:"params"` // JSON string with request/response params.
}

type apiParams struct {
	RequestParams []struct {
		Name     string `json:"name"`
		Required string `json:"required"`
		Type     string `json:"type"`
	} `json:"request_params"`
}

type catalog struct {
	Version   int         `json:"version"`
	Endpoints []catalogEP `json:"endpoints"`
}

type catalogEP struct {
	APIID      int            `json:"api_id"`
	APIName    string         `json:"api_name"`
	ModuleID   int            `json:"module_id"`
	ModuleName string         `json:"module_name"`
	Method     string         `json:"method"`
	Path       string         `json:"path"`
	DocURL     string         `json:"doc_url"`
	Request    []catalogParam `json:"request_params,omitempty"`
}

type catalogParam struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type,omitempty"`
}

func main() {
	var (
		baseURL = flag.String("base", "https://open.shopee.com/api/v1", "Shopee docs base URL")
		version = flag.Int("version", 2, "Docs version (1 or 2)")
		outPath = flag.String("out", filepath.FromSlash("shopee-sdk/catalog_v2.json"), "Output catalog path (relative to backend/)")
		sleepMs = flag.Int("sleep-ms", 30, "Sleep between API doc fetches (ms)")
	)
	flag.Parse()

	client := &http.Client{Timeout: 30 * time.Second}

	modulesURL := fmt.Sprintf("%s/doc/module/?version=%d", *baseURL, *version)
	var moduleResp moduleListResponse
	if err := getJSON(client, modulesURL, &moduleResp); err != nil {
		fatalf("fetch module list: %v", err)
	}

	var eps []catalogEP
	for _, m := range moduleResp.Modules {
		for _, it := range m.Items {
			if it.Type != 1 {
				continue
			}

			docURL := fmt.Sprintf("%s/doc/api/?api_id=%d", *baseURL, it.ID)
			var doc apiDoc
			if err := getJSON(client, docURL, &doc); err != nil {
				fmt.Fprintf(os.Stderr, "[WARN] api_id=%d (%s): %v\n", it.ID, it.Name, err)
				continue
			}

			method := "GET"
			if doc.Method == 1 {
				method = "POST"
			} else if doc.Method == 2 {
				method = "GET"
			}

			ep := catalogEP{
				APIID:      doc.APIID,
				APIName:    doc.APIName,
				ModuleID:   doc.ModuleID,
				ModuleName: doc.ModuleName,
				Method:     method,
				Path:       doc.Path,
				DocURL:     fmt.Sprintf("https://open.shopee.com/documents/v%d/%s?module=%d&type=1", *version, doc.APIName, doc.ModuleID),
			}

			if strings.TrimSpace(doc.Params) != "" {
				var p apiParams
				if err := json.Unmarshal([]byte(doc.Params), &p); err == nil {
					for _, rp := range p.RequestParams {
						ep.Request = append(ep.Request, catalogParam{
							Name:     rp.Name,
							Required: strings.EqualFold(rp.Required, "true"),
							Type:     rp.Type,
						})
					}
				}
			}

			eps = append(eps, ep)
			if *sleepMs > 0 {
				time.Sleep(time.Duration(*sleepMs) * time.Millisecond)
			}
		}
	}

	cat := catalog{Version: *version, Endpoints: eps}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fatalf("mkdir: %v", err)
	}
	f, err := os.Create(*outPath)
	if err != nil {
		fatalf("create output: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	// Keep it compressed so the embedded file stays small (one line).
	if err := enc.Encode(cat); err != nil {
		fatalf("write output: %v", err)
	}

	fmt.Printf("Wrote %d endpoints to %s\n", len(cat.Endpoints), *outPath)
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
