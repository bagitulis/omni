#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads - Executive Summary Section Builders
================================================
Helper functions untuk membangun section-section laporan.
"""

from datetime import datetime


def build_header(period_info):
    """Build report header section."""
    lines = [
        "# 📊 EXECUTIVE SUMMARY - TikTok Ads Performance",
        f"\n**Tanggal Laporan:** {datetime.now().strftime('%d %B %Y')}"
    ]
    if period_info:
        lines.append(f"\n**Periode Analisis:** {period_info.get('quarter_label', 'Rolling 3 Bulan')}")
    lines.append("\n---\n")
    return lines


def build_performance_summary(products, evaluation):
    """Build performance summary section."""
    health = evaluation.get('health', {})
    impact = evaluation.get('impact', {})
    dist = evaluation.get('distribution', {})
    
    current_profit = impact.get('current', {}).get('profit', 0)
    current_cost = impact.get('current', {}).get('cost', 0)
    projected_profit = impact.get('projected', {}).get('profit', 0)
    growth_pct = impact.get('projected', {}).get('growth_pct', 0)
    roi_overall = (current_profit + current_cost) / max(1, current_cost)
    
    total = sum([dist.get('lanjutkan', 0), dist.get('pantau', 0), dist.get('hentikan', 0)])
    active = len([p for p in products if p.get('is_still_active', True)])
    stopped = len(products) - active
    
    lines = [
        "## 💰 RINGKASAN KINERJA\n",
        "| Metrik | Nilai | % / Rasio |",
        "|--------|-------|-----------|",
        f"| **Health Score** | **{health.get('health_score', 0):.0f}/100** | Grade: {health.get('grade', 'N/A')} |",
        f"| Total Investasi | Rp {current_cost:,.0f} | 100% |",
        f"| Total Keuntungan | Rp {current_profit:,.0f} | ROI: {roi_overall:.1f}x |",
        f"| **Potensi Optimasi** | **Rp {projected_profit:,.0f}** | **+{growth_pct:.0f}%** |",
        "",
        "### Status Portfolio\n",
        f"**Total Produk:** {len(products)} | **Aktif:** {active} | **Sudah Di-stop:** {stopped}\n",
        f"- ✅ **LANJUTKAN:** {dist.get('lanjutkan', 0)} produk ({(dist.get('lanjutkan', 0) / max(1, total)) * 100:.0f}%)",
        f"- 🔶 **PANTAU:** {dist.get('pantau', 0)} produk ({(dist.get('pantau', 0) / max(1, total)) * 100:.0f}%)",
        f"- 🛑 **HENTIKAN:** {dist.get('hentikan', 0)} produk ({(dist.get('hentikan', 0) / max(1, total)) * 100:.0f}%)",
        "",
        "### 📐 Metodologi Statistik\n",
        "| Analisis | Metode | Threshold |",
        "|----------|--------|-----------|",
        "| Trend | Mann-Kendall Test | p < 0.05 |",
        "| Momentum | 70/30 Split Recent/Historical | ±20% |",
        "| Estimasi | ROI × 0.7 × TrendAdj × MomentumAdj | Conservative |",
        "| Confidence | HIGH (≥6 periode), MEDIUM (4-5), LOW (<4) | Data quality |",
        ""
    ]
    return lines


def build_scale_up_section(high_impact):
    """Build scale-up recommendation section."""
    lines = ["---\n", "## 🚀 REKOMENDASI: NAIKKAN BUDGET (Maks 10 Produk)\n",
             "*Produk dengan potensi tinggi jika budget ditambah 50%.*\n"]
    
    if not high_impact:
        lines.append("*Tidak ada produk yang memenuhi kriteria untuk scale-up.*\n")
        return lines
    
    video = [p for p in high_impact if p['creative_type'] == 'Video']
    kartu = [p for p in high_impact if p['creative_type'] == 'Kartu']
    
    lines.extend([
        f"**Total Tambahan Budget:** Rp {sum(p['budget_increase'] for p in high_impact):,.0f}",
        f"**Estimasi Tambahan Profit:** Rp {sum(p['estimated_additional_profit'] for p in high_impact):,.0f} (±15-30% margin error)\n"
    ])
    
    for ctype, items in [("🎬 Video", video), ("🃏 Kartu", kartu)]:
        if not items:
            continue
        lines.extend([f"### {ctype}\n",
            "| # | Produk | Budget | +Budget | Est. Profit | Confidence |",
            "|---|--------|--------|---------|-------------|------------|"])
        for i, p in enumerate(items, 1):
            lines.append(f"| {i} | {p['product_name'][:28]} | Rp {p['total_cost']:,.0f} | +Rp {p['budget_increase']:,.0f} | +Rp {p['estimated_additional_profit']:,.0f} | {p['confidence_label']} |")
        lines.append("")
    return lines


def build_maintain_section(maintain):
    """Build maintain budget section."""
    lines = ["---\n", "## ➡️ REKOMENDASI: PERTAHANKAN BUDGET\n",
             "*Produk dengan performa stabil, tidak perlu perubahan budget.*\n"]
    
    if not maintain:
        lines.append("*Tidak ada produk dalam kategori ini (semua masuk scale-up atau stop).*\n")
        return lines
    
    lines.extend([
        f"**Jumlah Produk:** {len(maintain)} | **Total Profit:** Rp {sum(p['profit'] for p in maintain):,.0f}\n",
        "| # | Produk | Tipe | ROI | Profit | Score | Confidence |",
        "|---|--------|------|-----|--------|-------|------------|"
    ])
    
    for i, p in enumerate(maintain[:10], 1):
        tipe = "🎬 Video" if p['creative_type'] == 'Video' else "🃏 Kartu"
        lines.append(f"| {i} | {p['product_name'][:28]} | {tipe} | {p['roi']:.1f}x | Rp {p['profit']:,.0f} | {p['score']:.0f} | {p['confidence_label']} |")
    
    if len(maintain) > 10:
        lines.append(f"| ... | *+{len(maintain)-10} produk lainnya* | | | | | |")
    lines.append("")
    return lines


def build_stop_section(stop_data):
    """Build stop products section."""
    active = stop_data['active']
    lines = ["---\n", "## 🛑 REKOMENDASI: HENTIKAN SEGERA (Masih Aktif)\n",
             "*Produk rugi yang masih berjalan dan perlu dihentikan.*\n"]
    
    if not active:
        lines.append("✅ *Tidak ada produk aktif yang perlu dihentikan.*\n")
        return lines
    
    lines.extend([
        f"**Jumlah Produk:** {len(active)}",
        f"**Kerugian Saat Ini:** Rp {sum(p['loss'] for p in active):,.0f}",
        f"**⚠️ Potensi Rugi Bulan Depan:** Rp {stop_data['total_potential_loss']:,.0f}\n"
    ])
    
    video = [p for p in active if p['creative_type'] == 'Video']
    kartu = [p for p in active if p['creative_type'] == 'Kartu']
    
    for ctype, items in [("🎬 Video", video), ("🃏 Kartu", kartu)]:
        if not items:
            continue
        lines.extend([f"### {ctype}\n",
            "| # | Produk | ROI | Kerugian | % Rugi | Potensi/Bulan | Confidence |",
            "|---|--------|-----|----------|--------|---------------|------------|"])
        for i, p in enumerate(items, 1):
            urgency = "🔴" if p['urgency'] == 'CRITICAL' else "🟠" if p['urgency'] == 'HIGH' else "🟡"
            lines.append(f"| {i} | {p['product_name'][:26]} | {p['roi']:.2f}x | Rp {p['loss']:,.0f} | {p['loss_pct']:.0f}% | Rp {p['potential_monthly_loss']:,.0f} | {urgency} {p['confidence_label']} |")
        lines.append("")
    return lines


def build_stopped_info_section(stop_data):
    """Build already stopped products info section."""
    stopped = stop_data['stopped']
    lines = ["---\n", "## ✅ PRODUK SUDAH DIHENTIKAN (Info)\n"]
    
    if not stopped:
        lines.append("*Tidak ada produk yang sudah di-stop.*\n")
        return lines
    
    lines.extend([
        f"**Jumlah:** {len(stopped)} | **Kerugian yang Dihindari:** Rp {sum(p['loss'] for p in stopped):,.0f}\n",
        "| # | Produk | Tipe | ROI | Kerugian Dihindari | Periode Terakhir |",
        "|---|--------|------|-----|-------------------|------------------|"
    ])
    
    for i, p in enumerate(stopped[:8], 1):
        tipe = "🎬 Video" if p['creative_type'] == 'Video' else "🃏 Kartu"
        last = p.get('last_active_period', 'N/A')[:15] if p.get('last_active_period') else 'N/A'
        lines.append(f"| {i} | {p['product_name'][:28]} | {tipe} | {p['roi']:.2f}x | Rp {p['loss']:,.0f} | {last} |")
    
    if len(stopped) > 8:
        lines.append(f"| ... | *+{len(stopped)-8} lainnya* | | | | |")
    lines.append("")
    return lines


def build_restart_section(restart_candidates):
    """Build restart recommendation section."""
    lines = ["---\n", "## 🔄 REKOMENDASI: AKTIFKAN KEMBALI\n",
             "*Produk yang sudah dihentikan tapi masih punya potensi bagus.*\n"]
    
    if not restart_candidates:
        lines.append("*Tidak ada produk yang direkomendasikan untuk restart.*\n")
        return lines
    
    lines.extend([
        f"**Jumlah Kandidat:** {len(restart_candidates)} | **Potensi Profit/Bulan:** Rp {sum(p['potential_monthly_profit'] for p in restart_candidates):,.0f}\n",
        "| # | Produk | Tipe | ROI Terakhir | Budget Rekomendasi | Potensi/Bulan | Confidence |",
        "|---|--------|------|--------------|-------------------|---------------|------------|"
    ])
    
    for i, p in enumerate(restart_candidates, 1):
        tipe = "🎬 Video" if p['creative_type'] == 'Video' else "🃏 Kartu"
        lines.append(f"| {i} | {p['product_name'][:26]} | {tipe} | {p['roi']:.1f}x | Rp {p['recommended_restart_budget']:,.0f}/minggu | Rp {p['potential_monthly_profit']:,.0f} | {p['confidence_label']} |")
    
    lines.extend(["", "**Catatan:** Mulai dengan 70% budget sebelumnya, pantau selama 2 minggu.\n"])
    return lines


def build_footer():
    """Build report footer."""
    return [
        "---\n",
        "### ⚠️ Catatan Confidence Level\n",
        "- **✅ HIGH:** Data lengkap (≥6 periode), rekomendasi dapat diandalkan",
        "- **🔶 MEDIUM:** Data cukup (4-5 periode), perlu monitoring lebih lanjut",
        "- **⚠️ LOW (RAGU):** Data terbatas (<4 periode), gunakan dengan hati-hati\n",
        "",
        f"*Executive Summary dibuat otomatis pada {datetime.now().strftime('%d %B %Y %H:%M')}*",
        "*TikTok Ads ML Analysis System v2.2.0 - Strict Statistical Methodology*"
    ]
