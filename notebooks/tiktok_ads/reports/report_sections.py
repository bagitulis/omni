#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Report Sections
======================================
Helper functions for generating individual report sections.
"""

from ..analysis import get_top_products, get_stop_products


def add_executive_summary(report, products, evaluation, period_info):
    """Add executive summary with health score"""
    report.append("## 🎯 RINGKASAN EKSEKUTIF\n")
    
    health = evaluation['health']
    risk = evaluation['risk']
    
    total_cost = sum(p['total_cost'] for p in products)
    total_revenue = sum(p['total_revenue'] for p in products)
    total_profit = sum(p['profit'] for p in products)
    avg_roi = total_revenue / total_cost if total_cost > 0 else 0
    
    if period_info:
        report.append(f"**Periode Analisis:** {period_info.get('start', 'N/A')} sampai {period_info.get('end', 'N/A')}")
        if period_info.get('rolling_window'):
            report.append(f" | **Metode:** Rolling Window {period_info.get('window_months', 3)} Bulan\n")
        else:
            report.append("\n")
    
    report.append("### 📊 Skor Kesehatan Portfolio\n")
    report.append(f"| Metrik | Nilai | Kategori |")
    report.append(f"|--------|-------|----------|")
    report.append(f"| **Health Score** | **{health['health_score']}/100** | **{health['grade']}** |")
    report.append(f"| Tingkat Risiko | {risk['risk_level']} | Skor: {risk['risk_score']}/100 |")
    report.append(f"| Status Keseluruhan | {health['interpretation']} | - |")
    report.append("")
    
    report.append("### 💰 Metrik Utama Performa\n")
    report.append(f"| Metrik | Nilai |")
    report.append(f"|--------|-------|")
    report.append(f"| Total Investasi Iklan | Rp {total_cost:,.0f} |")
    report.append(f"| Total Pendapatan | Rp {total_revenue:,.0f} |")
    report.append(f"| **Total Keuntungan Bersih** | **Rp {total_profit:,.0f}** |")
    report.append(f"| **ROI Keseluruhan** | **{avg_roi:.2f}x** |")
    report.append(f"| Jumlah Produk Dianalisis | {len(products)} |")
    report.append("")
    
    dist = evaluation['distribution']
    report.append("### 📈 Distribusi Rekomendasi Produk\n")
    report.append(f"| Rekomendasi | Jumlah Produk | Persentase |")
    report.append(f"|-------------|---------------|------------|")
    report.append(f"| 🟢 **LANJUTKAN** (Scale Up) | {dist['lanjutkan']} | {dist['lanjutkan']/len(products)*100:.1f}% |")
    report.append(f"| 🟡 **PANTAU** (Monitor) | {dist['pantau']} | {dist['pantau']/len(products)*100:.1f}% |")
    report.append(f"| 🔴 **HENTIKAN** (Stop) | {dist['hentikan']} | {dist['hentikan']/len(products)*100:.1f}% |")
    report.append("")


def add_ai_evaluation_section(report, evaluation):
    """Add comprehensive AI evaluation section"""
    report.append("## 🤖 EVALUASI BERBASIS AI & MACHINE LEARNING\n")
    
    health = evaluation['health']
    risk = evaluation['risk']
    
    report.append("### Komponen Health Score\n")
    report.append("Tabel berikut menunjukkan rincian perhitungan Health Score portfolio iklan:\n")
    report.append("| Komponen | Skor | Bobot | Nilai Aktual | Rumus Perhitungan |")
    report.append("|----------|------|-------|--------------|-------------------|")
    
    c = health['components']
    report.append(f"| Tingkat Profitabilitas | {c['profitability']['score']} | {c['profitability']['weight']} | {c['profitability']['value']} | (Produk Untung / Total) × 100 |")
    report.append(f"| Performa ROI | {c['roi_performance']['score']} | {c['roi_performance']['weight']} | {c['roi_performance']['value']} | 50 + log₁₀(ROI) × 30 |")
    report.append(f"| Kesehatan Trend | {c['trend_health']['score']} | {c['trend_health']['weight']} | {c['trend_health']['value']} | (Naik + Stabil×0.7) / Total |")
    report.append(f"| Kualitas Data | {c['data_quality']['score']} | {c['data_quality']['weight']} | {c['data_quality']['value']} | (Tinggi + Sedang×0.6) / Total |")
    report.append("")
    
    report.append("### ⚠️ Penilaian Risiko Portfolio\n")
    report.append(f"**Tingkat Risiko:** {risk['risk_level']} (Skor: {risk['risk_score']}/{risk['max_score']})\n")
    report.append("**Faktor-Faktor Risiko yang Teridentifikasi:**\n")
    for factor in risk['factors']:
        report.append(f"- {factor}")
    report.append("")


def add_profit_estimation_section(report, evaluation):
    """Add profit estimation if recommendations followed"""
    report.append("## 💹 PROYEKSI DAMPAK KEUANGAN\n")
    
    impact = evaluation['impact']
    if not impact:
        report.append("*Data tidak mencukupi untuk membuat proyeksi. Diperlukan lebih banyak data historis.*\n")
        return
    
    current = impact['current']
    projected = impact['projected']
    breakdown = impact['breakdown']
    assumptions = impact['assumptions']
    
    report.append("### Proyeksi Keuntungan Jika Rekomendasi Diterapkan\n")
    report.append("Tabel berikut menunjukkan estimasi peningkatan keuntungan berdasarkan analisis ML:\n")
    report.append(f"| Metrik | Nilai |")
    report.append(f"|--------|-------|")
    report.append(f"| Keuntungan Saat Ini | Rp {current['profit']:,.0f} |")
    report.append(f"| **Proyeksi Keuntungan** | **Rp {projected['profit']:,.0f}** |")
    report.append(f"| **Potensi Pertumbuhan** | **+{projected['growth_pct']:.1f}%** |")
    
    if 'confidence_interval' in projected:
        report.append(f"| **Confidence Interval 95%** | **{projected['confidence_interval']}** |")
    report.append("")
    
    conf_analysis = impact.get('confidence_analysis', {})
    if conf_analysis.get('is_valid'):
        report.append("### 📊 Analisis Tingkat Kepercayaan Statistik\n")
        report.append(f"| Metrik | Nilai | Penjelasan |")
        report.append(f"|--------|-------|------------|")
        report.append(f"| Confidence Level | {conf_analysis.get('level', '95%')} | Tingkat kepercayaan proyeksi |")
        report.append(f"| ROI Variability (CV) | {conf_analysis.get('roi_cv', 0):.1f}% | Variasi ROI antar produk |")
        report.append(f"| Uncertainty Factor | ±{conf_analysis.get('uncertainty_factor', 0):.1f}% | Rentang ketidakpastian estimasi |")
        report.append("")
    
    report.append("### Rincian Sumber Peningkatan Keuntungan\n")
    report.append(f"| Sumber Keuntungan | Nilai | Keterangan |")
    report.append(f"|-------------------|-------|------------|")
    report.append(f"| Penghematan dari Penghentian | +Rp {breakdown['savings_from_stop']:,.0f} | Menghentikan produk yang merugi |")
    report.append(f"| Keuntungan dari Scale-up | +Rp {breakdown['scale_up_gains']:,.0f} | Penambahan budget ke produk unggulan |")
    report.append(f"| Realokasi Budget | +Rp {breakdown['reallocation_potential']:,.0f} | Budget dialihkan ke produk dengan ROI tinggi |")
    report.append(f"| Budget yang Dibebaskan | Rp {breakdown['cost_freed']:,.0f} | Dari produk yang dihentikan |")
    report.append("")
    
    if impact.get('scale_up_details'):
        report.append("### Rincian Scale-Up Produk Unggulan\n")
        report.append("| Produk | Profit Saat Ini | Tambahan Budget | Estimasi Tambahan Profit | Confidence |")
        report.append("|--------|-----------------|-----------------|--------------------------|------------|")
        for d in impact['scale_up_details']:
            report.append(f"| {d['product']} | Rp {d['current_profit']:,.0f} | +Rp {d['additional_budget']:,.0f} | +Rp {d['projected_additional']:,.0f} | {d['confidence']:.0f}% |")
        report.append("")
    
    report.append("### Asumsi yang Digunakan dalam Perhitungan\n")
    report.append(f"- Persentase scale-up: {assumptions['scale_up_percentage']}")
    report.append(f"- Faktor konservasi ROI: {assumptions['roi_conservation_factor']} (pendekatan konservatif)")
    report.append(f"- Faktor ROI realokasi: {assumptions['reallocation_roi_factor']} (pendekatan konservatif)")
    report.append("")


def add_strategic_recommendations(report, products, evaluation):
    """Add strategic recommendations with methodology"""
    report.append("## 🎯 REKOMENDASI STRATEGIS BERBASIS DATA\n")
    
    for i, insight in enumerate(evaluation.get('strategic_insights', []), 1):
        report.append(f"### {i}. {insight['type']}\n")
        report.append(f"**Temuan Analisis:** {insight['insight']}\n")
        report.append(f"**Tindakan yang Disarankan:** {insight['action']}\n")
        report.append(f"**Metodologi yang Digunakan:** {insight['methodology']}\n")
        report.append("")
    
    _add_creative_type_comparison(report, evaluation)


def _add_creative_type_comparison(report, evaluation):
    """Add creative type comparison with T-Test"""
    cc = evaluation['creative_comparison']
    creative_stats = evaluation.get('creative_stats', {})
    
    if cc['kartu']['roi'] > cc['video']['roi']:
        better, worse = "Kartu", "Video"
        better_roi, worse_roi = cc['kartu']['roi'], cc['video']['roi']
    else:
        better, worse = "Video", "Kartu"
        better_roi, worse_roi = cc['video']['roi'], cc['kartu']['roi']
    
    diff_pct = ((better_roi / worse_roi) - 1) * 100 if worse_roi > 0 else 0
    
    report.append(f"### Strategi Berdasarkan Creative Type\n")
    report.append(f"| Tipe Creative | Jumlah Produk | ROI | Rekomendasi |")
    report.append(f"|---------------|---------------|-----|-------------|")
    report.append(f"| **{better}** | {cc['kartu']['count'] if better == 'Kartu' else cc['video']['count']} | **{better_roi:.2f}x** | ✅ Prioritaskan |")
    report.append(f"| {worse} | {cc['video']['count'] if worse == 'Video' else cc['kartu']['count']} | {worse_roi:.2f}x | Perlu Optimasi |")
    report.append(f"\n**Insight Performa:** Tipe {better} memiliki performa lebih baik dibandingkan {worse} sebesar **{diff_pct:.0f}%**\n")
    
    if creative_stats.get('is_valid'):
        _add_ttest_results(report, creative_stats)
    report.append("")


def _add_ttest_results(report, stats):
    """Add T-Test results table"""
    report.append("### 📊 Validasi Statistik Perbandingan Creative Type\n")
    report.append("Uji statistik untuk memvalidasi perbedaan performa antar creative type:\n")
    report.append(f"| Metrik Statistik | Video | Kartu |")
    report.append(f"|------------------|-------|-------|")
    report.append(f"| Ukuran Sampel (n) | {stats['video']['n']} | {stats['kartu']['n']} |")
    report.append(f"| Rata-rata ROI | {stats['video']['mean_roi']:.2f}x | {stats['kartu']['mean_roi']:.2f}x |")
    report.append(f"| Standar Deviasi | {stats['video']['std_roi']:.2f} | {stats['kartu']['std_roi']:.2f} |")
    report.append(f"| Median ROI | {stats['video']['median_roi']:.2f}x | {stats['kartu']['median_roi']:.2f}x |")
    report.append("")
    
    report.append(f"| Uji Statistik | Nilai Statistik | p-value | Signifikan? |")
    report.append(f"|---------------|-----------------|---------|-------------|")
    report.append(f"| T-Test | {stats['t_test']['statistic']:.3f} | {stats['t_test']['p_value']:.4f} | {'✅ Ya' if stats['t_test']['significant'] else '❌ Tidak'} |")
    report.append(f"| Mann-Whitney U | {stats['mann_whitney']['statistic']:.1f} | {stats['mann_whitney']['p_value']:.4f} | {'✅ Ya' if stats['mann_whitney']['significant'] else '❌ Tidak'} |")
    report.append("")
    
    overall = stats['overall']
    effect = stats['effect_size']
    report.append(f"**Kesimpulan Statistik:** {overall['significance']} {overall['significance_stars']}\n")
    report.append(f"**Ukuran Efek:** Cohen's d = {effect['cohens_d']:.3f} ({effect['interpretation']})\n")
    report.append(f"**Pemenang:** {stats['comparison']['winner']} ({stats['comparison']['winner_confidence']})\n")
