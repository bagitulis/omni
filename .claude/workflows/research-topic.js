// Workflow: Deep research on a topic with multiple sources
export const meta = {
  name: 'research-topic',
  description: 'Deep research dari web + codebase, cocok buat riset tech decision',
  phases: [
    { title: 'Search', detail: 'Cari info dari web dan codebase' },
    { title: 'Analyze', detail: 'Analisis dan sintesis temuan' },
    { title: 'Report', detail: 'Compile report dengan rekomendasi' },
  ],
}

const topic = args?.topic || 'current project architecture'

// Phase 1: Multi-modal search
phase('Search')
log(`Researching: ${topic}`)

const sources = await parallel([
  () => agent(
    `Search the web for recent best practices and trends about: ${topic}. Return key findings with URLs.`,
    { label: 'web-search', phase: 'Search' }
  ),
  () => agent(
    `Search the codebase for anything related to: ${topic}. Check docs/, README, ARCHITECTURE.md, and relevant code files.`,
    { label: 'codebase-search', phase: 'Search' }
  ),
  () => agent(
    `Find common pitfalls, anti-patterns, and lessons learned about: ${topic}. Search for blog posts, StackOverflow answers, and GitHub discussions.`,
    { label: 'pitfalls-search', phase: 'Search' }
  ),
])

log('All sources gathered')

// Phase 2: Cross-reference and analyze
phase('Analyze')
const crossRef = await agent(
  `Analyze these research findings. Cross-reference the web results with what's in the codebase.
   Identify: what's already good, what could improve, what are the risks.

   Web findings:
   ${sources[0]}

   Codebase findings:
   ${sources[1]}

   Pitfalls & lessons:
   ${sources[2]}`,
  { label: 'cross-reference', phase: 'Analyze' }
)

// Phase 3: Compile report
phase('Report')
const report = await agent(
  `Write a structured research report with:
   1. Executive Summary (3-5 sentences)
   2. Key Findings (bullet points)
   3. Recommendations (actionable, prioritized)
   4. Risks & Mitigations
   5. Sources

   Based on this analysis:
   ${crossRef}`,
  { label: 'write-report', phase: 'Report' }
)

return { report, raw_analysis: crossRef }
