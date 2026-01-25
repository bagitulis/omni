#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads - HTML Report Styles
================================
CSS styles for HTML reports with Shopee branding.
"""


def get_css_styles() -> str:
    """Return CSS styles for Shopee report with enhanced components."""
    return """
        :root {
            /* Shopee Brand Colors - Light Mode Optimized */
            --primary: #ee4d2d;
            --primary-light: #fff5f3;
            --primary-dark: #d4421e;
            --secondary: #26aa99;
            --secondary-light: #e6f7f5;
            --accent: #ff6633;
            
            /* Text Colors - WCAG AA Compliant (4.5:1 contrast) */
            --text-primary: #1a1a2e;
            --text-secondary: #374151;
            --text-muted: #6b7280;
            
            /* Background Colors */
            --bg-white: #ffffff;
            --bg-light: #f8fafc;
            --bg-section: #f1f5f9;
            
            /* Status Colors - High Contrast */
            --success: #166534;
            --success-bg: #dcfce7;
            --success-border: #22c55e;
            --warning: #92400e;
            --warning-bg: #fef3c7;
            --warning-border: #f59e0b;
            --danger: #991b1b;
            --danger-bg: #fee2e2;
            --danger-border: #ef4444;
            --info: #1e40af;
            --info-bg: #dbeafe;
            --info-border: #3b82f6;
            
            /* Borders & Shadows */
            --border: #e2e8f0;
            --shadow: 0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06);
            --shadow-lg: 0 4px 6px rgba(0,0,0,0.07), 0 2px 4px rgba(0,0,0,0.05);
            --shadow-xl: 0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -2px rgba(0,0,0,0.05);
        }
        
        * { box-sizing: border-box; margin: 0; padding: 0; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg-light);
            color: var(--text-primary);
            line-height: 1.7;
            padding: 24px;
            min-height: 100vh;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: var(--bg-white);
            border-radius: 12px;
            box-shadow: var(--shadow-lg);
            padding: 32px 40px;
        }
        
        /* Typography */
        h1 {
            color: var(--primary);
            font-size: 2rem;
            font-weight: 700;
            margin-bottom: 0.75rem;
            padding-bottom: 0.75rem;
            border-bottom: 3px solid var(--primary);
            letter-spacing: -0.025em;
        }
        
        h2 {
            color: var(--text-primary);
            font-size: 1.35rem;
            font-weight: 600;
            margin: 2.5rem 0 1rem;
            padding: 12px 16px;
            background: var(--bg-section);
            border-left: 4px solid var(--primary);
            border-radius: 0 8px 8px 0;
        }
        
        h3 {
            color: var(--text-primary);
            font-size: 1.1rem;
            font-weight: 600;
            margin: 1.75rem 0 1rem;
            padding-left: 12px;
            border-left: 3px solid var(--secondary);
        }
        
        /* KPI Cards Grid */
        .kpi-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 16px;
            margin: 20px 0;
        }
        
        .kpi-card {
            background: var(--bg-white);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 20px;
            text-align: center;
            box-shadow: var(--shadow);
            transition: transform 0.2s, box-shadow 0.2s;
        }
        
        .kpi-card:hover {
            transform: translateY(-2px);
            box-shadow: var(--shadow-xl);
        }
        
        .kpi-card.success { border-left: 4px solid var(--success-border); }
        .kpi-card.warning { border-left: 4px solid var(--warning-border); }
        .kpi-card.danger { border-left: 4px solid var(--danger-border); }
        .kpi-card.info { border-left: 4px solid var(--info-border); }
        .kpi-card.primary { border-left: 4px solid var(--primary); }
        
        .kpi-value {
            font-size: 1.75rem;
            font-weight: 700;
            color: var(--primary);
            margin-bottom: 4px;
        }
        
        .kpi-label {
            font-size: 0.85rem;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        
        /* Badge System */
        .badge {
            display: inline-block;
            padding: 4px 10px;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        
        .badge-success { background: var(--success-bg); color: var(--success); }
        .badge-warning { background: var(--warning-bg); color: var(--warning); }
        .badge-danger { background: var(--danger-bg); color: var(--danger); }
        .badge-info { background: var(--info-bg); color: var(--info); }
        .badge-primary { background: var(--primary-light); color: var(--primary); }
        
        /* Priority/Action Boxes */
        .priority-box {
            padding: 16px 20px;
            border-radius: 8px;
            margin: 16px 0;
            border-left: 4px solid;
        }
        
        .priority-high {
            background: var(--danger-bg);
            border-color: var(--danger-border);
            color: var(--danger);
        }
        
        .priority-medium {
            background: var(--warning-bg);
            border-color: var(--warning-border);
            color: var(--warning);
        }
        
        .priority-low {
            background: var(--success-bg);
            border-color: var(--success-border);
            color: var(--success);
        }
        
        .priority-info {
            background: var(--info-bg);
            border-color: var(--info-border);
            color: var(--info);
        }
        
        /* Health Score Meter */
        .health-meter {
            background: var(--bg-section);
            border-radius: 8px;
            padding: 20px;
            margin: 16px 0;
            text-align: center;
        }
        
        .health-score {
            font-size: 3rem;
            font-weight: 700;
            color: var(--primary);
        }
        
        .health-grade {
            font-size: 1.25rem;
            color: var(--text-secondary);
            margin-top: 8px;
        }
        
        /* Tables */
        table {
            width: 100%;
            border-collapse: separate;
            border-spacing: 0;
            margin: 16px 0;
            background: var(--bg-white);
            border-radius: 8px;
            overflow: hidden;
            box-shadow: var(--shadow);
            border: 1px solid var(--border);
        }
        
        th {
            background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
            color: white;
            padding: 14px 12px;
            text-align: left;
            font-weight: 600;
            font-size: 0.875rem;
        }
        
        td {
            padding: 12px;
            border-bottom: 1px solid var(--border);
            font-size: 0.875rem;
            color: var(--text-secondary);
        }
        
        tr:last-child td { border-bottom: none; }
        tr:hover { background: var(--primary-light); }
        tr:nth-child(even) { background: var(--bg-light); }
        tr:nth-child(even):hover { background: var(--primary-light); }
        
        /* Inline Styles */
        strong { color: var(--primary); font-weight: 600; }
        p { margin: 0.75rem 0; color: var(--text-secondary); }
        ul, ol { padding-left: 28px; margin: 12px 0; }
        li { margin: 6px 0; color: var(--text-secondary); }
        
        /* Blockquote - Info Box */
        blockquote {
            background: var(--info-bg);
            border-left: 4px solid var(--info);
            padding: 16px 20px;
            margin: 16px 0;
            border-radius: 0 8px 8px 0;
            color: var(--info);
            font-size: 0.9rem;
        }
        
        /* Code Blocks */
        code, pre {
            background: var(--bg-section);
            padding: 3px 8px;
            border-radius: 4px;
            font-family: 'SF Mono', 'Fira Code', monospace;
            font-size: 0.85rem;
            color: var(--text-primary);
            border: 1px solid var(--border);
        }
        
        pre {
            padding: 16px;
            overflow-x: auto;
            margin: 12px 0;
            line-height: 1.5;
        }
        
        pre code {
            border: none;
            padding: 0;
            background: transparent;
        }
        
        /* Horizontal Rule */
        hr {
            border: none;
            height: 1px;
            background: linear-gradient(90deg, transparent, var(--border), var(--primary), var(--border), transparent);
            margin: 32px 0;
        }
        
        /* Legend Section */
        .legend {
            background: var(--bg-section);
            border-radius: 8px;
            padding: 16px 20px;
            margin: 20px 0;
        }
        
        .legend-title {
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 12px;
        }
        
        .legend-items {
            display: flex;
            flex-wrap: wrap;
            gap: 16px;
        }
        
        .legend-item {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 0.85rem;
            color: var(--text-secondary);
        }
        
        /* Footer */
        .footer {
            margin-top: 48px;
            padding: 24px;
            text-align: center;
            border-top: 1px solid var(--border);
            color: var(--text-muted);
            font-size: 0.85rem;
            background: var(--bg-light);
            border-radius: 0 0 12px 12px;
            margin-left: -40px;
            margin-right: -40px;
            margin-bottom: -32px;
        }
        
        .footer p {
            margin: 4px 0;
            color: var(--text-muted);
        }
        
        /* Print Styles */
        @media print {
            body { background: white; padding: 0; }
            .container { box-shadow: none; padding: 20px; }
            h1 { font-size: 1.5rem; }
            h2 { font-size: 1.2rem; page-break-after: avoid; }
            table { box-shadow: none; page-break-inside: avoid; }
            .footer { background: white; }
            .kpi-card { box-shadow: none; border: 1px solid var(--border); }
        }
        
        /* Responsive */
        @media (max-width: 768px) {
            body { padding: 12px; }
            .container { padding: 20px; }
            h1 { font-size: 1.5rem; }
            h2 { font-size: 1.15rem; }
            table { font-size: 0.8rem; }
            th, td { padding: 8px 6px; }
            .kpi-grid { grid-template-columns: repeat(2, 1fr); }
            .kpi-value { font-size: 1.25rem; }
            .health-score { font-size: 2rem; }
        }
    """
