#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads - Reports Subpackage
===============================
Report generation and formatting modules.
"""

from .report import (
    generate_report_md,
    run_full_analysis,
    run_quarterly_analysis
)
from .report_html import generate_html
from .report_executive import generate_executive_summary

__all__ = [
    # Main report
    "generate_report_md",
    "run_full_analysis",
    "run_quarterly_analysis",
    # HTML
    "generate_html",
    # Executive
    "generate_executive_summary",
]
