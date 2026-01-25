package shopeesdk

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

//go:embed catalog_v2.json
var catalogFS embed.FS

type CatalogParam struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type,omitempty"`
}

type CatalogEndpoint struct {
	APIID      int    `json:"api_id"`
	APIName    string `json:"api_name"`
	ModuleID   int    `json:"module_id"`
	ModuleName string `json:"module_name"`
	Method     string `json:"method"` // "GET" or "POST"
	Path       string `json:"path"`   // "/api/v2/..."
	DocURL     string `json:"doc_url"`

	RequestParams []CatalogParam `json:"request_params,omitempty"`
}

type catalogFile struct {
	Version   int               `json:"version"`
	Endpoints []CatalogEndpoint `json:"endpoints"`
}

var (
	catalogOnce   sync.Once
	catalogErr    error
	catalogByName map[string]CatalogEndpoint
)

func loadCatalog() (map[string]CatalogEndpoint, error) {
	catalogOnce.Do(func() {
		data, err := catalogFS.ReadFile("catalog_v2.json")
		if err != nil {
			catalogErr = fmt.Errorf("read catalog_v2.json: %w", err)
			return
		}

		var cf catalogFile
		if err := json.Unmarshal(data, &cf); err != nil {
			catalogErr = fmt.Errorf("parse catalog_v2.json: %w", err)
			return
		}
		if cf.Version != 2 {
			catalogErr = fmt.Errorf("unexpected catalog version: %d", cf.Version)
			return
		}

		m := make(map[string]CatalogEndpoint, len(cf.Endpoints))
		for _, ep := range cf.Endpoints {
			if ep.APIName == "" || ep.Path == "" || ep.Method == "" {
				continue
			}
			m[ep.APIName] = ep
		}
		catalogByName = m
	})

	return catalogByName, catalogErr
}

// GetEndpoint returns endpoint metadata from the embedded official docs catalog.
func GetEndpoint(apiName string) (CatalogEndpoint, bool) {
	m, err := loadCatalog()
	if err != nil || m == nil {
		return CatalogEndpoint{}, false
	}
	ep, ok := m[apiName]
	return ep, ok
}

// CallAPI calls an API by its official docs name (e.g. "v2.product.get_recommend_attribute").
// Query params are appended to the signed URL; for POST, Body is encoded as JSON.
func (c *Client) CallAPI(ctx context.Context, apiName string, query map[string]string, body interface{}, out interface{}) error {
	if err := apiReady(ctx); err != nil {
		return err
	}
	if apiName == "" {
		return errors.New("api_name is required")
	}

	ep, ok := GetEndpoint(apiName)
	if !ok {
		return fmt.Errorf("unknown Shopee api_name: %s", apiName)
	}

	// Best-effort required param validation:
	// - GET: required params must exist in query.
	// - POST: if body is a map, required params must exist in body (or query).
	method := strings.ToUpper(ep.Method)
	switch method {
	case "GET":
		for _, p := range ep.RequestParams {
			if !p.Required {
				continue
			}
			if query == nil || strings.TrimSpace(query[p.Name]) == "" {
				return fmt.Errorf("missing required query param %q for %s", p.Name, apiName)
			}
		}
		return c.api.RawGet(ctx, ep.Path, query, out)
	case "POST":
		if bm, ok := body.(map[string]interface{}); ok {
			for _, p := range ep.RequestParams {
				if !p.Required {
					continue
				}
				_, inBody := bm[p.Name]
				inQuery := query != nil && strings.TrimSpace(query[p.Name]) != ""
				if !inBody && !inQuery {
					return fmt.Errorf("missing required param %q for %s", p.Name, apiName)
				}
			}
		}
		call := RawCall{
			Method: "POST",
			Path:   ep.Path,
			Query:  query,
			Body:   body,
		}
		return c.CallRaw(ctx, call, out)
	default:
		return fmt.Errorf("unsupported method %q for %s", ep.Method, apiName)
	}
}

// CallAPIMap calls an API by name and decodes into map[string]interface{}.
func (c *Client) CallAPIMap(ctx context.Context, apiName string, query map[string]string, body interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.CallAPI(ctx, apiName, query, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
