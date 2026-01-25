#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - HTML Report Generation
=============================================
Convert Markdown to professional HTML with TikTok styling.
"""

import re
from datetime import datetime


def generate_html(md_content, title="TikTok Ads Report"):
    """Convert Markdown to professional HTML with Light Mode styling"""
    html_body = _md_to_html(md_content)
    
    html = f"""<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{title}</title>
    <style>
        :root {{
            /* TikTok Brand Colors - Light Mode Optimized */
            --primary: #e62448;
            --primary-light: #fff0f3;
            --secondary: #00a1c9;
            --secondary-light: #e6f7fb;
            --accent: #5c6bc0;
            
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
            --warning: #92400e;
            --warning-bg: #fef3c7;
            --danger: #991b1b;
            --danger-bg: #fee2e2;
            --info: #1e40af;
            --info-bg: #dbeafe;
            
            /* Borders & Shadows */
            --border: #e2e8f0;
            --shadow: 0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06);
            --shadow-lg: 0 4px 6px rgba(0,0,0,0.07), 0 2px 4px rgba(0,0,0,0.05);
        }}
        
        * {{ box-sizing: border-box; margin: 0; padding: 0; }}
        
        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif;
            background: var(--bg-light);
            color: var(--text-primary);
            line-height: 1.7;
            padding: 24px;
            min-height: 100vh;
        }}
        
        .container {{
            max-width: 1200px;
            margin: 0 auto;
            background: var(--bg-white);
            border-radius: 12px;
            box-shadow: var(--shadow-lg);
            padding: 32px 40px;
        }}
        
        /* Typography */
        h1 {{
            color: var(--primary);
            font-size: 2rem;
            font-weight: 700;
            margin-bottom: 0.75rem;
            padding-bottom: 0.75rem;
            border-bottom: 3px solid var(--primary);
            letter-spacing: -0.025em;
        }}
        
        h2 {{
            color: var(--text-primary);
            font-size: 1.35rem;
            font-weight: 600;
            margin: 2.5rem 0 1rem;
            padding: 12px 16px;
            background: var(--bg-section);
            border-left: 4px solid var(--primary);
            border-radius: 0 8px 8px 0;
        }}
        
        h3 {{
            color: var(--text-primary);
            font-size: 1.1rem;
            font-weight: 600;
            margin: 1.75rem 0 1rem;
            padding-left: 12px;
            border-left: 3px solid var(--secondary);
        }}
        
        /* Tables - Professional & Clean */
        table {{
            width: 100%;
            border-collapse: separate;
            border-spacing: 0;
            margin: 16px 0;
            background: var(--bg-white);
            border-radius: 8px;
            overflow: hidden;
            box-shadow: var(--shadow);
            border: 1px solid var(--border);
        }}
        
        th {{
            background: linear-gradient(135deg, var(--primary) 0%, #c91d3e 100%);
            color: white;
            padding: 14px 12px;
            text-align: left;
            font-weight: 600;
            font-size: 0.875rem;
            text-transform: none;
            letter-spacing: 0.01em;
        }}
        
        td {{
            padding: 12px;
            border-bottom: 1px solid var(--border);
            font-size: 0.875rem;
            color: var(--text-secondary);
        }}
        
        tr:last-child td {{
            border-bottom: none;
        }}
        
        tr:hover {{
            background: var(--primary-light);
        }}
        
        tr:nth-child(even) {{
            background: var(--bg-light);
        }}
        
        tr:nth-child(even):hover {{
            background: var(--primary-light);
        }}
        
        /* Inline Styles */
        strong {{
            color: var(--primary);
            font-weight: 600;
        }}
        
        p {{
            margin: 0.75rem 0;
            color: var(--text-secondary);
        }}
        
        ul, ol {{
            padding-left: 28px;
            margin: 12px 0;
        }}
        
        li {{
            margin: 6px 0;
            color: var(--text-secondary);
        }}
        
        /* Blockquote - Info Box */
        blockquote {{
            background: var(--info-bg);
            border-left: 4px solid var(--info);
            padding: 16px 20px;
            margin: 16px 0;
            border-radius: 0 8px 8px 0;
            color: var(--info);
            font-size: 0.9rem;
        }}
        
        /* Code Blocks */
        code, pre {{
            background: var(--bg-section);
            padding: 3px 8px;
            border-radius: 4px;
            font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
            font-size: 0.85rem;
            color: var(--text-primary);
            border: 1px solid var(--border);
        }}
        
        pre {{
            padding: 16px;
            overflow-x: auto;
            margin: 12px 0;
            line-height: 1.5;
        }}
        
        pre code {{
            border: none;
            padding: 0;
            background: transparent;
        }}
        
        /* Horizontal Rule */
        hr {{
            border: none;
            height: 1px;
            background: linear-gradient(90deg, transparent, var(--border), var(--primary), var(--border), transparent);
            margin: 32px 0;
        }}
        
        /* Footer */
        .footer {{
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
        }}
        
        .footer p {{
            margin: 4px 0;
            color: var(--text-muted);
        }}
        
        /* Print Styles */
        @media print {{
            body {{
                background: white;
                padding: 0;
            }}
            .container {{
                box-shadow: none;
                padding: 20px;
            }}
            h1 {{ font-size: 1.5rem; }}
            h2 {{ font-size: 1.2rem; page-break-after: avoid; }}
            table {{ box-shadow: none; page-break-inside: avoid; }}
            .footer {{ background: white; }}
        }}
        
        /* Responsive */
        @media (max-width: 768px) {{
            body {{ padding: 12px; }}
            .container {{ padding: 20px; }}
            h1 {{ font-size: 1.5rem; }}
            h2 {{ font-size: 1.15rem; }}
            table {{ font-size: 0.8rem; }}
            th, td {{ padding: 8px 6px; }}
        }}
    </style>
</head>
<body>
<div class="container">
{html_body}
<div class="footer">
    <p><strong>TikTok Ads ML Analysis System v2.2.0</strong></p>
    <p>Dibuat pada: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
    <p style="margin-top: 8px; font-size: 0.8rem;">Metodologi: Mann-Kendall Test • Linear Regression • Momentum Analysis • Coefficient of Variation</p>
</div>
</div>
</body>
</html>"""
    
    return html


def _md_to_html(md_content):
    """Enhanced Markdown to HTML conversion"""
    lines = md_content.split('\n')
    html_lines = []
    in_table = False
    in_list = False
    in_code = False
    list_type = None
    
    for line in lines:
        if line.strip() == '```':
            if in_code:
                html_lines.append("</pre>")
                in_code = False
            else:
                html_lines.append("<pre>")
                in_code = True
            continue
        
        if in_code:
            html_lines.append(line)
            continue
        
        if line.startswith('# '):
            html_lines.append(f"<h1>{line[2:]}</h1>")
        elif line.startswith('## '):
            html_lines.append(f"<h2>{line[3:]}</h2>")
        elif line.startswith('### '):
            html_lines.append(f"<h3>{line[4:]}</h3>")
        elif line.startswith('> '):
            html_lines.append(f"<blockquote>{line[2:]}</blockquote>")
        elif line.strip() == '---':
            html_lines.append("<hr>")
        elif line.startswith('|'):
            if not in_table:
                html_lines.append("<table>")
                in_table = True
            
            if '---' in line:
                continue
            
            cells = [c.strip() for c in line.split('|')[1:-1]]
            if len(html_lines) > 0 and html_lines[-1] == "<table>":
                html_lines.append("<thead><tr>" + "".join(f"<th>{c}</th>" for c in cells) + "</tr></thead><tbody>")
            else:
                html_lines.append("<tr>" + "".join(f"<td>{c}</td>" for c in cells) + "</tr>")
        else:
            if in_table:
                html_lines.append("</tbody></table>")
                in_table = False
            
            if line.startswith('- '):
                if not in_list or list_type != 'ul':
                    if in_list:
                        html_lines.append(f"</{list_type}>")
                    html_lines.append("<ul>")
                    in_list = True
                    list_type = 'ul'
                html_lines.append(f"<li>{line[2:]}</li>")
            elif re.match(r'^\d+\. ', line):
                if not in_list or list_type != 'ol':
                    if in_list:
                        html_lines.append(f"</{list_type}>")
                    html_lines.append("<ol>")
                    in_list = True
                    list_type = 'ol'
                html_lines.append(f"<li>{re.sub(r'^\\d+\\. ', '', line)}</li>")
            else:
                if in_list:
                    html_lines.append(f"</{list_type}>")
                    in_list = False
                    list_type = None
                
                line = re.sub(r'\*\*(.+?)\*\*', r'<strong>\1</strong>', line)
                line = re.sub(r'\*(.+?)\*', r'<em>\1</em>', line)
                line = re.sub(r'`(.+?)`', r'<code>\1</code>', line)
                
                if line.strip():
                    html_lines.append(f"<p>{line}</p>")
    
    if in_table:
        html_lines.append("</tbody></table>")
    if in_list:
        html_lines.append(f"</{list_type}>")
    if in_code:
        html_lines.append("</pre>")
    
    return "\n".join(html_lines)
