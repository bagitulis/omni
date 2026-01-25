#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Report Sections (Advanced)
=================================================
Correlation, Churn Risk, Budget, and additional sections.
"""

from ..analysis import get_top_products, get_stop_products


def add_correlation_section(report, evaluation):
    """Add correlation analysis section"""
    report.append("## 📈 ANALISIS KORELASI & EFISIENSI PENGELUARAN\n")
    
    corr = evaluation.get('correlation_analysis', {})
    if not corr.get('is_valid'):
        report.append("*Data tidak mencukupi untuk melakukan analisis korelasi.*\n")
        return
    
    cost_rev = corr['cost_revenue']
    cost_roi = corr['cost_roi']
    efficiency = corr['efficiency_analysis']
    
    report.append("### Korelasi antara Biaya dan Pendapatan\n")
    report.append(f"| Metrik | Nilai | Interpretasi |")
    report.append(f"|--------|-------|--------------|")
    report.append(f"| Pearson r | {cost_rev['pearson_r']:.3f} | {cost_rev['interpretation']} |")
    report.append(f"| p-value | {cost_rev['pearson_p']:.4f} | {'✅ Signifikan secara statistik' if cost_rev['significant'] else '❌ Tidak signifikan'} |")
    report.append(f"| Spearman r | {cost_rev['spearman_r']:.3f} | Validasi non-parametrik |")
    report.append("")
    
    report.append("### Analisis Diminishing Returns\n")
    report.append(f"| Metrik | Nilai | Interpretasi |")
    report.append(f"|--------|-------|--------------|")
    report.append(f"| Korelasi Biaya-ROI | {cost_roi['correlation']:.3f} | {cost_roi['interpretation']} |")
    report.append(f"| Diminishing Returns? | {'⚠️ Ya - ROI menurun saat pengeluaran naik' if cost_roi['diminishing_returns'] else '✅ Tidak - Pengeluaran masih efektif'} | {'Perlu batasi budget' if cost_roi['diminishing_returns'] else 'Aman untuk scale up'} |")
    report.append("")
    
    if efficiency.get('efficient_products'):
        eff_products = efficiency['efficient_products']
        video_eff = [p for p in eff_products if p.get('creative_type') == 'Video']
        kartu_eff = [p for p in eff_products if p.get('creative_type') == 'Kartu']
        
        report.append("### 🎯 Produk dengan Efisiensi Tinggi (Kandidat Scale Up)\n")
        report.append("*Pengeluaran rendah dengan ROI tinggi - potensi yang belum dimanfaatkan secara maksimal*\n")
        
        if video_eff:
            report.append("#### 🎬 Produk Video yang Efisien\n")
            report.append(f"| Produk | Pengeluaran Saat Ini | ROI | Peluang |")
            report.append(f"|--------|----------------------|-----|---------|")
            for p in video_eff[:5]:
                report.append(f"| {p['product_name'][:35]} | Rp {p['cost']:,.0f} | {p['roi']:.1f}x | Pengeluaran rendah, ROI tinggi - scale up |")
            report.append("")
        
        if kartu_eff:
            report.append("#### 🃏 Produk Kartu yang Efisien\n")
            report.append(f"| Produk | Pengeluaran Saat Ini | ROI | Peluang |")
            report.append(f"|--------|----------------------|-----|---------|")
            for p in kartu_eff[:5]:
                report.append(f"| {p['product_name'][:35]} | Rp {p['cost']:,.0f} | {p['roi']:.1f}x | Pengeluaran rendah, ROI tinggi - scale up |")
            report.append("")
    
    if efficiency.get('inefficient_products'):
        ineff_products = efficiency['inefficient_products']
        video_ineff = [p for p in ineff_products if p.get('creative_type') == 'Video']
        kartu_ineff = [p for p in ineff_products if p.get('creative_type') == 'Kartu']
        
        report.append("### ⚠️ Produk dengan Efisiensi Rendah (Perlu Evaluasi)\n")
        report.append("*Pengeluaran tinggi dengan ROI di bawah rata-rata - perlu optimasi atau pengurangan budget*\n")
        
        if video_ineff:
            report.append("#### 🎬 Produk Video yang Kurang Efisien\n")
            report.append(f"| Produk | Pengeluaran Saat Ini | ROI | Masalah |")
            report.append(f"|--------|----------------------|-----|---------|")
            for p in video_ineff[:5]:
                report.append(f"| {p['product_name'][:35]} | Rp {p['cost']:,.0f} | {p['roi']:.1f}x | Pengeluaran tinggi, ROI di bawah rata-rata |")
            report.append("")
        
        if kartu_ineff:
            report.append("#### 🃏 Produk Kartu yang Kurang Efisien\n")
            report.append(f"| Produk | Pengeluaran Saat Ini | ROI | Masalah |")
            report.append(f"|--------|----------------------|-----|---------|")
            for p in kartu_ineff[:5]:
                report.append(f"| {p['product_name'][:35]} | Rp {p['cost']:,.0f} | {p['roi']:.1f}x | Pengeluaran tinggi, ROI di bawah rata-rata |")
            report.append("")
    
    for ins in corr.get('insights', []):
        report.append(f"- **[{ins['type']}]** {ins['insight']}")
        report.append(f"  - *Tindakan:* {ins['action']}")
    report.append("")


def add_churn_risk_section(report, evaluation):
    """Add churn risk / early warning section - SEPARATED by Creative Type"""
    report.append("## ⚠️ SISTEM PERINGATAN DINI (Churn Risk)\n")
    report.append("Analisis risiko penurunan performa produk berdasarkan momentum, trend, dan volatilitas.\n")
    
    churn = evaluation.get('churn_analysis', {})
    if not churn.get('is_valid'):
        report.append("*Data tidak mencukupi untuk melakukan analisis risiko.*\n")
        return
    
    dist = churn['distribution']
    
    report.append("### Gambaran Umum Risiko Portfolio\n")
    report.append(f"| Metrik | Nilai |")
    report.append(f"|--------|-------|")
    report.append(f"| Skor Risiko Portfolio | **{churn['portfolio_risk_score']:.1f}/100** |")
    report.append(f"| Produk Risiko Tinggi | {dist['high_risk']} ({churn['high_risk_pct']:.1f}%) |")
    report.append(f"| Produk Risiko Sedang | {dist['medium_risk']} |")
    report.append(f"| Produk Risiko Rendah | {dist['low_risk']} |")
    report.append("")
    
    warnings = churn.get('early_warnings', [])
    if warnings:
        # Separate by creative type
        video_warnings = [w for w in warnings if w.get('creative_type') == 'Video']
        kartu_warnings = [w for w in warnings if w.get('creative_type') == 'Kartu']
        
        if video_warnings:
            report.append("### 🎬 Produk Video dengan Risiko Tinggi\n")
            report.append("| # | Produk | Skor Risiko | Tingkat | Prediksi |")
            report.append("|---|--------|-------------|---------|----------|")
            for i, w in enumerate(video_warnings[:5], 1):
                name = w.get('product_name', w.get('name', 'N/A'))[:30]
                report.append(f"| {i} | {name} | {w['risk_score']}/100 | {w['risk_level']} | Kemungkinan tinggi mengalami penurunan |")
            report.append("")
        
        if kartu_warnings:
            report.append("### 🃏 Produk Kartu dengan Risiko Tinggi\n")
            report.append("| # | Produk | Skor Risiko | Tingkat | Prediksi |")
            report.append("|---|--------|-------------|---------|----------|")
            for i, w in enumerate(kartu_warnings[:5], 1):
                name = w.get('product_name', w.get('name', 'N/A'))[:30]
                report.append(f"| {i} | {name} | {w['risk_score']}/100 | {w['risk_level']} | Kemungkinan tinggi mengalami penurunan |")
            report.append("")
        
        report.append("### Rincian Faktor Risiko (3 Produk Teratas)\n")
        for w in warnings[:3]:
            ctype = w.get('creative_type', 'N/A')
            emoji = "🎬" if ctype == "Video" else "🃏"
            name = w.get('product_name', w.get('name', 'N/A'))[:30]
            report.append(f"**{emoji} [{ctype}] {name}** (Skor: {w['risk_score']})\n")
            for f in w.get('risk_factors', []):
                report.append(f"- [{f['severity']}] {f['factor']}: {f['value']} (+{f['points']} poin)")
            report.append("")
    else:
        report.append("✅ Tidak ada produk dengan risiko tinggi saat ini. Portfolio dalam kondisi baik.\n")


def add_budget_optimization_section(report, evaluation):
    """Add budget optimization recommendations - SEPARATED by Creative Type"""
    report.append("## 💰 REKOMENDASI OPTIMASI BUDGET\n")
    report.append("Alokasi budget optimal berdasarkan historical ROI, momentum, dan composite score.\n")
    
    budget = evaluation.get('budget_optimization', {})
    if not budget.get('is_valid'):
        report.append("*Data tidak mencukupi untuk memberikan rekomendasi optimasi budget.*\n")
        return
    
    summary = budget['summary']
    
    report.append("### Ringkasan Rekomendasi Budget\n")
    report.append(f"| Metrik | Nilai |")
    report.append(f"|--------|-------|")
    report.append(f"| Total Budget Saat Ini | Rp {budget['current_total_budget']:,.0f} |")
    report.append(f"| Total Budget yang Direkomendasikan | Rp {budget['recommended_total_budget']:,.0f} |")
    report.append(f"| Budget yang Dapat Dibebaskan (dari STOP) | Rp {budget['budget_freed']:,.0f} |")
    report.append(f"| Estimasi Dampak Pendapatan | **+Rp {budget['estimated_revenue_impact']:,.0f}** |")
    report.append("")
    
    report.append("### Ringkasan Tindakan\n")
    report.append(f"| Tindakan | Jumlah Produk |")
    report.append(f"|----------|---------------|")
    report.append(f"| 🛑 HENTIKAN | {summary['products_to_stop']} |")
    report.append(f"| 📉 KURANGI | {summary['products_to_reduce']} |")
    report.append(f"| ➡️ PERTAHANKAN | {summary['products_to_maintain']} |")
    report.append(f"| 📈 TINGKATKAN | {summary['products_to_increase']} |")
    report.append("")
    
    top_inc = budget.get('top_increases', [])
    if top_inc:
        # Separate by creative type
        video_inc = [r for r in top_inc if r.get('creative_type') == 'Video']
        kartu_inc = [r for r in top_inc if r.get('creative_type') == 'Kartu']
        
        if video_inc:
            report.append("### 🎬 Produk Video untuk Scale-Up\n")
            report.append("| Produk | Saat Ini | Direkomendasikan | Perubahan | Est. Dampak |")
            report.append("|--------|----------|------------------|-----------|-------------|")
            for r in video_inc[:5]:
                final = r.get('final_recommended_budget', r['recommended_budget'])
                change = final - r['current_budget']
                if change > 0:
                    report.append(f"| {r['product_name'][:30]} | Rp {r['current_budget']:,.0f} | Rp {final:,.0f} | +Rp {change:,.0f} | +Rp {r['estimated_revenue_impact']:,.0f} |")
            report.append("")
        
        if kartu_inc:
            report.append("### 🃏 Produk Kartu untuk Scale-Up\n")
            report.append("| Produk | Saat Ini | Direkomendasikan | Perubahan | Est. Dampak |")
            report.append("|--------|----------|------------------|-----------|-------------|")
            for r in kartu_inc[:5]:
                final = r.get('final_recommended_budget', r['recommended_budget'])
                change = final - r['current_budget']
                if change > 0:
                    report.append(f"| {r['product_name'][:30]} | Rp {r['current_budget']:,.0f} | Rp {final:,.0f} | +Rp {change:,.0f} | +Rp {r['estimated_revenue_impact']:,.0f} |")
            report.append("")
    
    to_stop = budget.get('to_stop', [])
    if to_stop:
        video_stop = [r for r in to_stop if r.get('creative_type') == 'Video']
        kartu_stop = [r for r in to_stop if r.get('creative_type') == 'Kartu']
        
        if video_stop:
            report.append("### 🛑 Produk Video untuk DIHENTIKAN\n")
            report.append("| Produk | Budget Saat Ini | Alasan |")
            report.append("|--------|-----------------|--------|")
            for r in video_stop[:5]:
                report.append(f"| {r['product_name'][:30]} | Rp {r['current_budget']:,.0f} | {r['rationale'][:40]} |")
            report.append("")
        
        if kartu_stop:
            report.append("### 🛑 Produk Kartu untuk DIHENTIKAN\n")
            report.append("| Produk | Budget Saat Ini | Alasan |")
            report.append("|--------|-----------------|--------|")
            for r in kartu_stop[:5]:
                report.append(f"| {r['product_name'][:30]} | Rp {r['current_budget']:,.0f} | {r['rationale'][:40]} |")
            report.append("")
    
    report.append("> **Catatan Metodologi:** Rekomendasi budget optimal dihitung berdasarkan historical ROI, momentum performa, dan composite score masing-masing produk.\n")


def add_data_quality_section(report, products):
    """Add data quality section"""
    report.append("## 📊 KUALITAS DATA ANALISIS\n")
    report.append("Tingkat kepercayaan analisis bergantung pada kualitas dan kelengkapan data yang tersedia.\n")
    
    high = len([p for p in products if '✅' in p['data_quality']])
    medium = len([p for p in products if '🔶' in p['data_quality']])
    low = len([p for p in products if '⚠️' in p['data_quality']])
    insufficient = len([p for p in products if '❌' in p['data_quality']])
    n = len(products)
    
    report.append(f"| Kualitas Data | Jumlah | Persentase | Keterangan |")
    report.append(f"|---------------|--------|------------|------------|")
    report.append(f"| ✅ TINGGI | {high} | {high/n*100:.0f}% | ≥8 periode data, sangat reliable |")
    report.append(f"| 🔶 SEDANG | {medium} | {medium/n*100:.0f}% | 4-7 periode data, cukup reliable |")
    report.append(f"| ⚠️ RENDAH | {low} | {low/n*100:.0f}% | <4 periode data, perlu hati-hati |")
    report.append(f"| ❌ TIDAK CUKUP | {insufficient} | {insufficient/n*100:.0f}% | <2 periode data, tidak valid untuk analisis |")
    report.append("")
    report.append("> **Catatan:** Rekomendasi untuk produk dengan kualitas data RENDAH/TIDAK CUKUP ditandai dengan `[DATA TERBATAS]` dan perlu evaluasi tambahan.\n")


def add_top_products_section(report, products):
    """Add top products with statistical details - SEPARATED by Creative Type"""
    report.append("## 🏆 PRODUK UNGGULAN - REKOMENDASI SCALE UP\n")
    report.append("Daftar produk dengan performa terbaik yang disarankan untuk ditingkatkan budget iklannya.\n")
    
    top = get_top_products(products, n=20)
    if not top:
        report.append("*Tidak ada produk yang memenuhi kriteria produk unggulan.*\n")
        return
    
    # Separate by creative type
    top_video = [p for p in top if p['creative_type'] == 'Video']
    top_kartu = [p for p in top if p['creative_type'] == 'Kartu']
    
    # VIDEO section
    if top_video:
        report.append("### 🎬 Produk Video Unggulan\n")
        report.append("| # | Produk | ROI | Keuntungan | Skor | Trend | Momentum |")
        report.append("|---|--------|-----|------------|------|-------|----------|")
        for i, p in enumerate(top_video[:10], 1):
            report.append(f"| {i} | {p['product_name'][:35]} | {p['roi']:.1f}x | Rp {p['profit']:,.0f} | {p['score']:.1f} | {p['trend']} | {p['momentum']} |")
        report.append("")
    
    # KARTU section
    if top_kartu:
        report.append("### 🃏 Produk Kartu Unggulan\n")
        report.append("| # | Produk | ROI | Keuntungan | Skor | Trend | Momentum |")
        report.append("|---|--------|-----|------------|------|-------|----------|")
        for i, p in enumerate(top_kartu[:10], 1):
            report.append(f"| {i} | {p['product_name'][:35]} | {p['roi']:.1f}x | Rp {p['profit']:,.0f} | {p['score']:.1f} | {p['trend']} | {p['momentum']} |")
        report.append("")
    
    report.append("### Kriteria Produk Unggulan\n")
    report.append("Produk dikategorikan sebagai **UNGGULAN** jika memenuhi **SEMUA** kriteria berikut:\n")
    report.append("1. **Composite Score ≥ 60** - Skor gabungan dari ROI (30%), Profit (25%), Momentum (20%), Konsistensi (15%), Trend (10%)")
    report.append("2. **ROI ≥ 2.0x** - Pendapatan minimal 2 kali lipat dari biaya iklan")
    report.append("3. **Profit > 0** - Produk harus menghasilkan keuntungan (tidak boleh merugi)")
    report.append("")


def add_stop_products_section(report, products):
    """Add stop products with justification - SEPARATED by Creative Type and Activity Status"""
    report.append("## 🚫 PRODUK UNTUK DIHENTIKAN\n")
    report.append("Daftar produk yang disarankan untuk dihentikan iklannya karena performa yang buruk.\n")
    
    stop = get_stop_products(products)
    if not stop:
        report.append("✅ *Tidak ada produk yang perlu dihentikan. Semua produk memiliki performa yang memadai.*\n")
        return
    
    # Separate by activity status
    active_stop = [p for p in stop if p.get('is_still_active', True)]
    already_stopped = [p for p in stop if not p.get('is_still_active', True)]
    
    total_loss = sum(abs(p['profit']) for p in stop if p['profit'] < 0)
    active_loss = sum(abs(p['profit']) for p in active_stop if p['profit'] < 0)
    
    report.append(f"**Total Produk Bermasalah:** {len(stop)} | **Masih Aktif:** {len(active_stop)} | **Sudah Di-stop:** {len(already_stopped)}")
    report.append(f"**Total Kerugian:** Rp {total_loss:,.0f} | **Kerugian Aktif:** Rp {active_loss:,.0f}\n")
    
    # === ACTIVE PRODUCTS THAT NEED TO STOP ===
    if active_stop:
        report.append("### ⚠️ PERLU AKSI: Produk Masih Aktif\n")
        report.append("*Produk berikut masih aktif dan perlu dihentikan segera:*\n")
        
        # Separate by creative type
        stop_video = [p for p in active_stop if p['creative_type'] == 'Video']
        stop_kartu = [p for p in active_stop if p['creative_type'] == 'Kartu']
        
        # VIDEO to STOP
        if stop_video:
            video_loss = sum(abs(p['profit']) for p in stop_video if p['profit'] < 0)
            video_loss_pct = (video_loss / max(1, sum(p['total_cost'] for p in stop_video))) * 100
            report.append(f"#### 🎬 Produk Video Aktif ({len(stop_video)} produk, kerugian: Rp {video_loss:,.0f} = {video_loss_pct:.0f}%)\n")
            report.append("| # | Produk | ROI | Kerugian | % Rugi | Skor | Status |")
            report.append("|---|--------|-----|----------|--------|------|--------|")
            for i, p in enumerate(stop_video[:10], 1):
                loss = abs(p['profit']) if p['profit'] < 0 else 0
                loss_pct = (loss / max(1, p['total_cost'])) * 100
                report.append(f"| {i} | {p['product_name'][:32]} | {p['roi']:.2f}x | Rp {loss:,.0f} | {loss_pct:.0f}% | {p['score']:.1f} | ⚠️ AKTIF |")
            report.append("")
        
        # KARTU to STOP
        if stop_kartu:
            kartu_loss = sum(abs(p['profit']) for p in stop_kartu if p['profit'] < 0)
            kartu_loss_pct = (kartu_loss / max(1, sum(p['total_cost'] for p in stop_kartu))) * 100
            report.append(f"#### 🃏 Produk Kartu Aktif ({len(stop_kartu)} produk, kerugian: Rp {kartu_loss:,.0f} = {kartu_loss_pct:.0f}%)\n")
            report.append("| # | Produk | ROI | Kerugian | % Rugi | Skor | Status |")
            report.append("|---|--------|-----|----------|--------|------|--------|")
            for i, p in enumerate(stop_kartu[:10], 1):
                loss = abs(p['profit']) if p['profit'] < 0 else 0
                loss_pct = (loss / max(1, p['total_cost'])) * 100
                report.append(f"| {i} | {p['product_name'][:32]} | {p['roi']:.2f}x | Rp {loss:,.0f} | {loss_pct:.0f}% | {p['score']:.1f} | ⚠️ AKTIF |")
            report.append("")
    
    # === ALREADY STOPPED PRODUCTS (INFORMATIONAL) ===
    if already_stopped:
        report.append("### ✅ SUDAH DIHENTIKAN (Informasi)\n")
        report.append("*Produk berikut sudah tidak aktif di periode terakhir:*\n")
        
        stopped_loss = sum(abs(p['profit']) for p in already_stopped if p['profit'] < 0)
        report.append(f"**Jumlah:** {len(already_stopped)} produk | **Kerugian Dihindari:** Rp {stopped_loss:,.0f}\n")
        
        report.append("| # | Produk | Tipe | ROI | Kerugian | Periode Terakhir |")
        report.append("|---|--------|------|-----|----------|-----------------|")
        for i, p in enumerate(already_stopped[:10], 1):
            loss = abs(p['profit']) if p['profit'] < 0 else 0
            emoji = "🎬" if p['creative_type'] == 'Video' else "🃏"
            last_period = p.get('last_active_period', 'N/A')
            report.append(f"| {i} | {p['product_name'][:32]} | {emoji} | {p['roi']:.2f}x | Rp {loss:,.0f} | {last_period} |")
        
        if len(already_stopped) > 10:
            report.append(f"| ... | *+{len(already_stopped) - 10} produk lainnya* | | | | |")
        report.append("")
    
    if not active_stop and not already_stopped:
        report.append("✅ *Tidak ada produk yang perlu dihentikan.*\n")
        report.append("")
    
    report.append("### Kriteria Penghentian Produk\n")
    report.append("Produk direkomendasikan untuk **DIHENTIKAN** jika memenuhi **SALAH SATU** kriteria berikut:\n")
    report.append("1. **Composite Score < 40** - Skor keseluruhan terlalu rendah")
    report.append("2. **ROI < 1.0x** - Pendapatan lebih rendah dari biaya (merugi)")
    report.append("3. **Profit < -Rp 100.000** - Kerugian melebihi batas toleransi")
    report.append("")


def add_creative_type_section(report, products):
    """Add creative type comparison"""
    report.append("## 📱 PERBANDINGAN PERFORMA CREATIVE TYPE\n")
    report.append("Analisis perbandingan performa antara creative type Video dan Kartu.\n")
    
    for ctype in ["Video", "Kartu"]:
        filtered = [p for p in products if p['creative_type'] == ctype]
        if not filtered:
            continue
        
        cost = sum(p['total_cost'] for p in filtered)
        revenue = sum(p['total_revenue'] for p in filtered)
        profit = sum(p['profit'] for p in filtered)
        roi = revenue / cost if cost > 0 else 0
        profitable = len([p for p in filtered if p['profit'] > 0])
        
        emoji = "🎬" if ctype == "Video" else "🃏"
        report.append(f"### {emoji} Performa {ctype}\n")
        report.append(f"| Metrik | Nilai |")
        report.append(f"|--------|-------|")
        report.append(f"| Jumlah Produk | {len(filtered)} |")
        report.append(f"| Produk Profitable | {profitable} ({profitable/len(filtered)*100:.0f}%) |")
        report.append(f"| Total Biaya Iklan | Rp {cost:,.0f} |")
        report.append(f"| Total Pendapatan | Rp {revenue:,.0f} |")
        report.append(f"| **Total Keuntungan** | **Rp {profit:,.0f}** |")
        report.append(f"| **ROI Keseluruhan** | **{roi:.2f}x** |")
        report.append("")
