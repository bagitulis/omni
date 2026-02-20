# Frontend Refactoring Plan (RULES.md Compliance)

## Target: Files > 300 lines (violating ~300 Lines rule)

### Priority 1: API Layer (types + functions mixed → separate)
| # | File | Lines | Strategy |
|---|------|-------|----------|
| 1 | `api/wholesale.ts` | 486 | Split types → `types/wholesale.ts`, helpers → `api/wholesaleHelpers.ts` |
| 2 | `api/client.ts` | 383 | Extract response types → `types/api.ts`, extract operation methods → `api/operations.ts` |
| 3 | `api/orders.ts` | 356 | Extract types → `types/order.ts` (if not already), split bulk ops → `api/orderBulkOps.ts` |

### Priority 2: Page Components (logic + JSX mixed → hook extraction)
| # | File | Lines | Strategy |
|---|------|-------|----------|
| 4 | `pages/settings/tabs/GoogleSheetsTab.tsx` | 362 | Extract hook → `hooks/useGoogleSheetsConfig.ts`, extract sub-components |
| 5 | `pages/inventory/components/InventoryHeader.tsx` | 319 | Extract column settings popover → separate component |
| 6 | `pages/route-mapping/components/GraphView.tsx` | 319 | Extract custom nodes → `GraphNodes.tsx` |
| 7 | `pages/route-mapping/components/GraphViewHelpers.ts` | 309 | Already a helper — extract D3 layout → `graphLayout.ts` |
| 8 | `pages/inventory/SimplifiedInventoryPage.tsx` | 300 | Extract tab items config, sync handlers already in hooks |

### Priority 3: Shared Components (large render functions → sub-components)
| # | File | Lines | Strategy |
|---|------|-------|----------|
| 9 | `components/modals/PriceModal.tsx` | 333 | Extract preview table → `PricePreviewTable.tsx`, step config |
| 10 | `components/shared/SkuMappingPanel.tsx` | 323 | Extract link modal → `SkuLinkModal.tsx`, platform cell → separate |
| 11 | `components/modals/InventoryColumnsModal.tsx` | 301 | Extract group builder logic and column group renderer |
| 12 | `components/shared/InlineEditCell.tsx` | 300 | Extract display/edit mode → sub-components |

## Rules Compliance Checklist
- [ ] SRP: One function = one purpose
- [ ] DRY: No duplicate logic, extract to utilities
- [ ] OOP: Proper encapsulation, use interfaces
- [ ] ~300 Lines: All files under ~300 lines after refactoring
