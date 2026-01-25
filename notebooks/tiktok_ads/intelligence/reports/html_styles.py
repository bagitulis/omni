#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
HTML Styles Module
==================
CSS styles untuk HTML reports.
Dipisah dari main generator untuk SRP compliance.
"""


CSS_STYLE = """
<style>
    :root {
        --primary: #e62448;
        --primary-light: #fff0f3;
        --secondary: #00a1c9;
        --text-primary: #1a1a2e;
        --text-secondary: #374151;
        --text-muted: #6b7280;
        --bg-white: #ffffff;
        --bg-light: #f8fafc;
        --bg-section: #f1f5f9;
        --success: #166534;
        --success-bg: #dcfce7;
        --warning: #92400e;
        --warning-bg: #fef3c7;
        --danger: #991b1b;
        --danger-bg: #fee2e2;
        --info: #1e40af;
        --info-bg: #dbeafe;
        --border: #e2e8f0;
        --shadow: 0 1px 3px rgba(0,0,0,0.1);
        --shadow-lg: 0 4px 6px rgba(0,0,0,0.07);
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        background: var(--bg-light);
        color: var(--text-primary);
        line-height: 1.7;
        padding: 24px;
    }
    .container {
        max-width: 1200px;
        margin: 0 auto;
        background: var(--bg-white);
        border-radius: 12px;
        box-shadow: var(--shadow-lg);
        padding: 32px 40px;
    }
    h1 {
        color: var(--primary);
        font-size: 2rem;
        font-weight: 700;
        margin-bottom: 0.75rem;
        padding-bottom: 0.75rem;
        border-bottom: 3px solid var(--primary);
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
        background: linear-gradient(135deg, var(--primary) 0%, #c91d3e 100%);
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
        vertical-align: top;
    }
    /* Product name cell - allow wrapping */
    td:nth-child(2) {
        min-width: 180px;
        max-width: 250px;
        line-height: 1.4;
    }
    td small {
        display: block;
        color: var(--text-muted);
        font-size: 0.75rem;
        margin-top: 2px;
    }
    tr:last-child td { border-bottom: none; }
    tr:hover { background: var(--primary-light); }
    tr:nth-child(even) { background: var(--bg-light); }
    strong { color: var(--primary); font-weight: 600; }
    p { margin: 0.75rem 0; color: var(--text-secondary); }
    ul, ol { padding-left: 28px; margin: 12px 0; }
    li { margin: 6px 0; color: var(--text-secondary); }
    blockquote {
        background: var(--info-bg);
        border-left: 4px solid var(--info);
        padding: 16px 20px;
        margin: 16px 0;
        border-radius: 0 8px 8px 0;
        color: var(--info);
    }
    pre {
        background: var(--bg-section);
        padding: 16px;
        border-radius: 8px;
        overflow-x: auto;
        font-family: 'SF Mono', monospace;
        font-size: 0.85rem;
    }
    hr {
        border: none;
        height: 1px;
        background: linear-gradient(90deg, transparent, var(--border), var(--primary), var(--border), transparent);
        margin: 32px 0;
    }
    .footer {
        margin-top: 48px;
        padding: 24px;
        text-align: center;
        border-top: 1px solid var(--border);
        color: var(--text-muted);
        font-size: 0.85rem;
        background: var(--bg-light);
    }
    .badge-success { background: var(--success-bg); color: var(--success); padding: 2px 8px; border-radius: 4px; }
    .badge-warning { background: var(--warning-bg); color: var(--warning); padding: 2px 8px; border-radius: 4px; }
    .badge-danger { background: var(--danger-bg); color: var(--danger); padding: 2px 8px; border-radius: 4px; }
    
    /* KPI Cards Grid */
    .kpi-grid {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 16px;
        margin: 20px 0;
    }
    .kpi-card {
        padding: 20px;
        border-radius: 12px;
        text-align: center;
        box-shadow: var(--shadow);
        border: 1px solid var(--border);
    }
    .kpi-card.kpi-success { background: var(--success-bg); border-color: var(--success); }
    .kpi-card.kpi-warning { background: var(--warning-bg); border-color: var(--warning); }
    .kpi-card.kpi-danger { background: var(--danger-bg); border-color: var(--danger); }
    .kpi-card.kpi-info { background: var(--info-bg); border-color: var(--info); }
    .kpi-value { font-size: 1.75rem; font-weight: 700; margin-bottom: 4px; }
    .kpi-card.kpi-success .kpi-value { color: var(--success); }
    .kpi-card.kpi-warning .kpi-value { color: var(--warning); }
    .kpi-card.kpi-danger .kpi-value { color: var(--danger); }
    .kpi-card.kpi-info .kpi-value { color: var(--info); }
    .kpi-label { font-size: 0.85rem; color: var(--text-muted); margin-bottom: 4px; }
    .kpi-status { font-size: 0.8rem; font-weight: 600; }
    
    /* Priority Boxes */
    .priority-box {
        padding: 16px 20px;
        border-radius: 8px;
        margin: 12px 0;
        border-left: 5px solid;
    }
    .priority-critical { background: var(--danger-bg); border-color: var(--danger); }
    .priority-high { background: var(--warning-bg); border-color: var(--warning); }
    .priority-normal { background: var(--success-bg); border-color: var(--success); }
    .priority-badge {
        font-weight: 700;
        font-size: 0.9rem;
        margin-bottom: 8px;
    }
    .priority-critical .priority-badge { color: var(--danger); }
    .priority-high .priority-badge { color: var(--warning); }
    .priority-normal .priority-badge { color: var(--success); }
    .priority-content { color: var(--text-secondary); font-size: 0.9rem; }
    .priority-content ul { margin: 8px 0 0 0; padding-left: 20px; }
    .priority-content li { margin: 4px 0; }
    
    /* Event Timeline */
    .event-timeline {
        display: flex;
        gap: 12px;
        margin: 16px 0;
        flex-wrap: wrap;
    }
    .event-item {
        padding: 10px 16px;
        border-radius: 8px;
        font-size: 0.85rem;
        font-weight: 500;
    }
    .event-today { background: var(--primary-light); color: var(--primary); border: 1px solid var(--primary); }
    .event-upcoming { background: var(--warning-bg); color: var(--warning); border: 1px solid var(--warning); }
    .event-future { background: var(--info-bg); color: var(--info); border: 1px solid var(--info); }
    
    /* Section Type Headers */
    .section-video { border-left-color: #9333ea !important; }
    .section-card { border-left-color: #0891b2 !important; }
    h4.type-video { color: #9333ea; }
    h4.type-card { color: #0891b2; }
    
    /* Discontinued Notice */
    .discontinued-notice {
        background: var(--bg-section);
        padding: 12px 16px;
        border-radius: 8px;
        color: var(--text-muted);
        font-size: 0.85rem;
        margin: 16px 0;
    }
    
    /* Badge Styles */
    .badge {
        display: inline-block;
        padding: 3px 8px;
        border-radius: 4px;
        font-size: 0.75rem;
        font-weight: 600;
        white-space: nowrap;
        cursor: help;
    }
    .badge-success { background: var(--success-bg); color: var(--success); }
    .badge-warning { background: var(--warning-bg); color: var(--warning); }
    .badge-danger { background: var(--danger-bg); color: var(--danger); }
    .badge-info { background: var(--info-bg); color: var(--info); }
    .badge-growth { background: #d1fae5; color: #065f46; }
    .badge-mature { background: #dbeafe; color: #1e40af; }
    .badge-decline { background: #fee2e2; color: #991b1b; }
    .badge-stable { background: #f3f4f6; color: #374151; }
    
    /* CI Range Small Text */
    td small { color: var(--text-muted); font-size: 0.75rem; }
    
    /* Legend Section */
    .legend-section {
        background: var(--bg-section);
        border-radius: 12px;
        padding: 24px;
        margin: 24px 0;
    }
    .legend-group {
        margin-bottom: 20px;
        padding-bottom: 16px;
        border-bottom: 1px solid var(--border);
    }
    .legend-group:last-child {
        margin-bottom: 0;
        padding-bottom: 0;
        border-bottom: none;
    }
    .legend-group h4 {
        color: var(--text-primary);
        font-size: 1rem;
        font-weight: 600;
        margin-bottom: 12px;
    }
    .legend-group ul {
        list-style: none;
        padding: 0;
        margin: 0;
    }
    .legend-group li {
        padding: 6px 0;
        font-size: 0.9rem;
        color: var(--text-secondary);
        display: flex;
        align-items: center;
        gap: 8px;
    }
    .legend-group li .badge {
        min-width: 120px;
        text-align: center;
    }
</style>
"""
