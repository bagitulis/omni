// Simple check that imports work
import('src/pages/analytics/MLDashboardPage.tsx').then(() => console.log('ML Dashboard imports OK')).catch((e) => console.error('Import failed:', e.message));
