// TikTok Ads MCP Server
package main

import (
	"fmt"

	"mcp-servers/pkg/mcp"
	"mcp-servers/pkg/python"
)

var runner = python.NewRunner("tiktok_ads")

func main() {
	server := mcp.NewServer("tiktok-ads-analyzer", "2.1.0")

	server.RegisterTools([]mcp.Tool{
		{
			Name:        "get_menu",
			Description: "TRIGGER: 'MCP Menu TikTok' or 'MCP TikTok' - Shows available TikTok Ads tools",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "analyze_ads",
			Description: "TRIGGER: 'MCP Analisis TikTok' - Get summary statistics (cost, revenue, ROI)",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_top_products",
			Description: "TRIGGER: 'MCP Produk Terbaik TikTok' - Get top products by ML score",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"limit": {Type: "number", Description: "Number of products (default: 10)"},
				},
				Required: []string{},
			},
		},
		{
			Name:        "get_stop_products",
			Description: "TRIGGER: 'MCP Produk Stop TikTok' - Get products to stop advertising",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "generate_insights",
			Description: "TRIGGER: 'MCP Insight TikTok' - Generate insights with ML analysis and recommendations",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "generate_report",
			Description: "TRIGGER: 'MCP Laporan TikTok' - Generate full HTML/MD report",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
	})

	server.SetHandler(func(name string, args map[string]interface{}) (interface{}, error) {
		switch name {
		case "get_menu":
			return getMenu(), nil
		case "analyze_ads":
			return analyzeAds()
		case "get_top_products":
			limit := 10
			if l, ok := args["limit"].(float64); ok {
				limit = int(l)
			}
			return getTopProducts(limit)
		case "get_stop_products":
			return getStopProducts()
		case "generate_insights":
			return generateInsights()
		case "generate_report":
			return generateFullReport()
		default:
			return nil, fmt.Errorf("unknown tool: %s", name)
		}
	})

	server.Run()
}

func getMenu() map[string]interface{} {
	return map[string]interface{}{
		"title":       "🎵 MCP TikTok Ads Menu",
		"description": "Gunakan perintah 'MCP + [Trigger]' untuk mengaktifkan tool",
		"tools": []map[string]string{
			{"trigger": "MCP Menu TikTok", "label": "📋 Menu", "description": "Tampilkan daftar perintah ini"},
			{"trigger": "MCP Analisis TikTok", "label": "📊 Analisis Ringkas", "description": "Lihat statistik umum ads"},
			{"trigger": "MCP Produk Terbaik TikTok", "label": "🏆 Produk Terbaik", "description": "Produk performa terbaik"},
			{"trigger": "MCP Produk Stop TikTok", "label": "🛑 Produk Stop", "description": "Produk harus dihentikan"},
			{"trigger": "MCP Insight TikTok", "label": "💡 Insight Lengkap", "description": "Analisis mendalam"},
			{"trigger": "MCP Laporan TikTok", "label": "📑 Generate Laporan", "description": "Buat laporan HTML"},
		},
	}
}

func analyzeAds() (interface{}, error) {
	return runner.RunCode(`
import json
from tiktok_ads import load_data, summary_stats

df = load_data(verbose=False)
stats = summary_stats(df)

by_mode = df.groupby('bidding_mode').agg({
    'product_id': 'nunique', 'cost': 'sum', 'revenue': 'sum'
}).reset_index()
by_mode['roi'] = by_mode['revenue'] / by_mode['cost']

result = {
    'summary': {
        'totalProducts': int(stats['total_products']),
        'totalCost': int(stats['total_cost']),
        'totalRevenue': int(stats['total_revenue']),
        'totalProfit': int(stats['total_profit']),
        'overallRoi': round(stats['roi'], 2),
        'totalRecords': int(stats['total_records'])
    },
    'byBiddingMode': [
        {'mode': r['bidding_mode'], 'products': int(r['product_id']),
         'cost': int(r['cost']), 'revenue': int(r['revenue']), 'roi': round(r['roi'], 2)}
        for _, r in by_mode.iterrows()
    ]
}
print(json.dumps(result))
`)
}

func getTopProducts(limit int) (interface{}, error) {
	return runner.RunCode(fmt.Sprintf(`
import json
from tiktok_ads import load_data, analyze_products, get_top_products

df = load_data(verbose=False)
products = analyze_products(df, verbose=False)
top = get_top_products(products, n=%d)

result = [
    {'productId': p['product_id'], 'productName': p['product_name'],
     'biddingMode': p['bidding_mode'], 'totalCost': p['total_cost'],
     'totalRevenue': p['total_revenue'], 'profit': p['profit'],
     'roi': p['roi'], 'score': p['score'], 'trend': p['trend'],
     'momentum': p['momentum'], 'category': p['category']}
    for p in top
]
print(json.dumps(result))
`, limit))
}

func getStopProducts() (interface{}, error) {
	return runner.RunCode(`
import json
from tiktok_ads import load_data, analyze_products, get_stop_products

df = load_data(verbose=False)
products = analyze_products(df, verbose=False)
stop = get_stop_products(products)

result = [
    {'productId': p['product_id'], 'productName': p['product_name'],
     'biddingMode': p['bidding_mode'], 'profit': p['profit'],
     'roi': p['roi'], 'score': p['score'], 'category': p['category'],
     'action': p['action']}
    for p in stop
]
print(json.dumps(result))
`)
}

func generateInsights() (interface{}, error) {
	return runner.RunCode(`
import json
from tiktok_ads import load_data, analyze_products, summary_stats, get_top_products, get_stop_products

df = load_data(verbose=False)
products = analyze_products(df, verbose=False)
stats = summary_stats(df)

top5 = get_top_products(products, n=5)
stop5 = get_stop_products(products)[:5]
savings = sum(abs(p['profit']) for p in products if p['profit'] < 0)

lanjut = len([p for p in products if 'LANJUTKAN' in p['category']])
pantau = len([p for p in products if 'PANTAU' in p['category']])
henti = len([p for p in products if 'HENTIKAN' in p['category']])

by_mode = df.groupby('bidding_mode').agg({'product_id': 'nunique', 'cost': 'sum', 'revenue': 'sum'}).reset_index()
by_mode['roi'] = by_mode['revenue'] / by_mode['cost']

def fmt(n): return f"Rp {n:,.0f}".replace(',', '.')

result = {
    'overview': {
        'status': 'SANGAT BAIK' if stats['roi'] > 5 else 'BAIK' if stats['roi'] > 2 else 'PERLU PERHATIAN',
        'totalInvestment': fmt(stats['total_cost']),
        'totalRevenue': fmt(stats['total_revenue']),
        'totalProfit': fmt(stats['total_profit']),
        'roi': f"{stats['roi']:.2f}x",
        'distribution': {'LANJUTKAN': lanjut, 'PANTAU': pantau, 'HENTIKAN': henti}
    },
    'recommendations': {
        'scaleUp': [{'product': p['product_name'], 'roi': f"{p['roi']}x", 'score': p['score'],
                     'trend': p['trend'], 'action': 'Tambah budget 30-50%'} for p in top5],
        'stop': [{'product': p['product_name'], 'loss': fmt(abs(p['profit'])),
                  'roi': f"{p['roi']}x", 'action': 'STOP segera'} for p in stop5],
        'potentialSavings': fmt(savings)
    },
    'biddingComparison': [
        {'mode': r['bidding_mode'], 'products': int(r['product_id']),
         'cost': int(r['cost']), 'revenue': int(r['revenue']), 'roi': round(r['roi'], 2)}
        for _, r in by_mode.iterrows()
    ]
}
print(json.dumps(result))
`)
}

func generateFullReport() (interface{}, error) {
	return runner.RunCode(`
import json
from tiktok_ads.reports import run_quarterly_analysis

result = run_quarterly_analysis(verbose=False)
def fmt(n): return f"Rp {n:,.0f}".replace(',', '.')

evaluation = result.get('evaluation', {})
dist = evaluation.get('distribution', {})

output = {
    'status': 'success',
    'message': 'Laporan berhasil dibuat',
    'summary': {
        'status': evaluation.get('status', 'N/A'),
        'totalProducts': evaluation.get('total_products', 0),
        'roi': f"{evaluation.get('roi', 0):.2f}x",
        'distribution': dist
    },
    'files': result.get('files', {})
}
print(json.dumps(output))
`)
}
