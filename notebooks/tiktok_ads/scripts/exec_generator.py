#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads - Executive Summary Generator
=========================================
Generate Executive Summary markdown report.
Refactored for AGENTS.MD compliance (max 300 lines).
"""

from .exec_categorizers import (
    get_high_impact_products, get_potential_restart_products,
    get_maintain_budget_products, get_all_stop_products_categorized
)
from .exec_sections import (
    build_header, build_performance_summary, build_scale_up_section,
    build_maintain_section, build_stop_section, build_stopped_info_section,
    build_restart_section, build_footer
)


def _build_action_items(high_impact, maintain, stop_data, restart_candidates, dist):
    """Build action items section."""
    active_stop = stop_data['active']
    lines = ["---\n", "## 📋 ACTION ITEMS MINGGU INI\n"]
    action_num = 1
    
    if active_stop:
        total_potential = stop_data['total_potential_loss']
        priority = '🔴 URGENT' if total_potential > 500000 else '🟠 HIGH'
        lines.extend([
            f"**{action_num}. 🛑 HENTIKAN {len(active_stop)} PRODUK RUGI**",
            f"   - Kerugian saat ini: Rp {sum(p['loss'] for p in active_stop):,.0f}",
            f"   - Potensi kerugian jika dilanjut: Rp {total_potential:,.0f}/bulan",
            f"   - Prioritas: {priority}", ""
        ])
        action_num += 1
    
    if high_impact:
        lines.extend([
            f"**{action_num}. 🚀 NAIKKAN BUDGET {len(high_impact)} PRODUK UNGGULAN**",
            f"   - Tambahan budget: Rp {sum(p['budget_increase'] for p in high_impact):,.0f} (+50%)",
            f"   - Estimasi profit tambahan: Rp {sum(p['estimated_additional_profit'] for p in high_impact):,.0f}", ""
        ])
        action_num += 1
    
    if restart_candidates:
        lines.extend([
            f"**{action_num}. 🔄 AKTIFKAN KEMBALI {len(restart_candidates)} PRODUK**",
            f"   - Potensi profit: Rp {sum(p['potential_monthly_profit'] for p in restart_candidates):,.0f}/bulan",
            f"   - Mulai dengan 70% budget sebelumnya", ""
        ])
        action_num += 1
    
    if maintain:
        lines.extend([
            f"**{action_num}. ➡️ PERTAHANKAN {len(maintain)} PRODUK STABIL**",
            f"   - Total profit saat ini: Rp {sum(p['profit'] for p in maintain):,.0f}", ""
        ])
        action_num += 1
    
    pantau = dist.get('pantau', 0)
    if pantau > 0:
        lines.extend([f"**{action_num}. 🔶 MONITOR {pantau} PRODUK**", f"   - Evaluasi ulang minggu depan", ""])
    
    return lines


def generate_executive_summary_md(products, evaluation, period_info=None):
    """Generate comprehensive Executive Summary with all recommendations."""
    # Get categorized data
    high_impact = get_high_impact_products(products, max_products=10)
    high_impact_ids = {(p['product_name'], p['creative_type']) for p in high_impact}
    
    maintain = get_maintain_budget_products(products, high_impact_ids)
    stop_data = get_all_stop_products_categorized(products)
    restart_candidates = get_potential_restart_products(products)
    dist = evaluation.get('distribution', {})
    
    # Build report sections
    report = []
    report.extend(build_header(period_info))
    report.extend(build_performance_summary(products, evaluation))
    report.extend(build_scale_up_section(high_impact))
    report.extend(build_maintain_section(maintain))
    report.extend(build_stop_section(stop_data))
    report.extend(build_stopped_info_section(stop_data))
    report.extend(build_restart_section(restart_candidates))
    report.extend(_build_action_items(high_impact, maintain, stop_data, restart_candidates, dist))
    report.extend(build_footer())
    
    return "\n".join(report)
