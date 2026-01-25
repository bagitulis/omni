#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Analysis - HTML Report Generation
=============================================
Convert Markdown to professional HTML with Shopee styling.
Refactored: CSS in report_styles.py, converter in report_converter.py
"""

from datetime import datetime
from .report_styles import get_css_styles
from .report_converter import md_to_html


def generate_html(md_content: str, title: str = "Shopee Ads Report") -> str:
    """Convert Markdown to professional HTML with Shopee Light Mode styling."""
    html_body = md_to_html(md_content)
    css_styles = get_css_styles()
    
    html = f"""<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{title}</title>
    <style>
{css_styles}
    </style>
</head>
<body>
<div class="container">
{html_body}
<div class="footer">
    <p><strong>Shopee Ads ML Analysis System v2.1.0</strong></p>
    <p>Dibuat pada: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
    <p style="margin-top: 8px; font-size: 0.8rem;">Metodologi: Unified Scorer (8 Components) • Mann-Kendall Test • Budget Elasticity • Churn Risk • Indonesian Calendar</p>
</div>
</div>
</body>
</html>"""
    
    return html
