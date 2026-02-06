# Fungsi untuk mendapatkan versi terbaru
function Get-LatestVersion {
    param($packageName)
    try {
        $version = npm view $packageName version 2>$null
        if ($version) {
            return $version
        }
    }
    catch {
        Write-Warning "Gagal mendapatkan versi untuk $packageName"
    }
    return $null
}

# Daftar package untuk diupdate
$packages = @{
    # Dependencies
    "axios" = Get-LatestVersion "axios"
    "eslint-plugin-vue" = Get-LatestVersion "eslint-plugin-vue"
    "pinia" = Get-LatestVersion "pinia"
    "vue" = Get-LatestVersion "vue"
    "vue-eslint-parser" = Get-LatestVersion "vue-eslint-parser"
    "vue-router" = Get-LatestVersion "vue-router"
    
    # DevDependencies
    "@tailwindcss/forms" = Get-LatestVersion "@tailwindcss/forms"
    "@tailwindcss/typography" = Get-LatestVersion "@tailwindcss/typography"
    "@typescript-eslint/eslint-plugin" = Get-LatestVersion "@typescript-eslint/eslint-plugin"
    "@typescript-eslint/parser" = Get-LatestVersion "@typescript-eslint/parser"
    "@vitejs/plugin-vue" = Get-LatestVersion "@vitejs/plugin-vue"
    "@vue/tsconfig" = Get-LatestVersion "@vue/tsconfig"
    "autoprefixer" = Get-LatestVersion "autoprefixer"
    "daisyui" = Get-LatestVersion "daisyui"
    "eslint" = Get-LatestVersion "eslint"
    "postcss" = Get-LatestVersion "postcss"
    "sharp" = Get-LatestVersion "sharp"
    "tailwindcss" = Get-LatestVersion "tailwindcss"
    "terser" = Get-LatestVersion "terser"
    "ts-node" = Get-LatestVersion "ts-node"
    "tsx" = Get-LatestVersion "tsx"
    "typescript" = Get-LatestVersion "typescript"
    "vite" = Get-LatestVersion "vite"
}

# Tampilkan versi terbaru
Write-Host "`nVersi terbaru yang ditemukan:" -ForegroundColor Green
$packages.GetEnumerator() | Sort-Object Name | Format-Table -AutoSize

# Update package.json
$packageJson = Get-Content package.json -Raw | ConvertFrom-Json

# Update dependencies
foreach ($pkg in $packages.GetEnumerator()) {
    if ($pkg.Value) {
        if ($packageJson.dependencies.PSObject.Properties.Name -contains $pkg.Name) {
            $packageJson.dependencies.$($pkg.Name) = "^$($pkg.Value)"
        }
        elseif ($packageJson.devDependencies.PSObject.Properties.Name -contains $pkg.Name) {
            $packageJson.devDependencies.$($pkg.Name) = "^$($pkg.Value)"
        }
    }
}

# Simpan package.json
$packageJson | ConvertTo-Json -Depth 10 | Set-Content package.json
Write-Host "`npackage.json berhasil diupdate!" -ForegroundColor Green

# Hapus node_modules dan install ulang
Write-Host "`nMenghapus node_modules dan package-lock.json..." -ForegroundColor Yellow
Remove-Item -Recurse -Force node_modules, package-lock.json -ErrorAction SilentlyContinue

Write-Host "`nMenginstall dependencies..." -ForegroundColor Yellow
npm install

Write-Host "`nSelesai! Versi terbaru telah diinstall." -ForegroundColor Green