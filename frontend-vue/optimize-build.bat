@echo off
REM Frontend optimization and build script for Windows

echo 🚀 Starting frontend build optimization...

REM Install dependencies
call npm install

REM Clean previous build
if exist dist (
    rmdir /s /q dist
)

REM Build with optimizations
call npm run build

echo.
echo ✅ Build optimization complete!
echo.
echo Performance improvements applied:
echo   ✓ Terser minification (passes: 4, aggressive compression)
echo   ✓ CSS code splitting and minification
echo   ✓ Route-based code splitting
echo   ✓ Platform-specific lazy loading
echo   ✓ Removed console.log and debugger statements
echo   ✓ Fixed back/forward cache compatibility
echo   ✓ Optimized resource loading with requestIdleCallback
echo.
echo To view build report:
echo   npm run build:report
