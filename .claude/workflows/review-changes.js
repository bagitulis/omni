// Workflow: Review git changes across multiple dimensions
export const meta = {
  name: 'review-changes',
  description: 'Review git diff dari berbagai angle: bugs, security, performance',
  phases: [
    { title: 'Gather', detail: 'Kumpulin diff yang berubah' },
    { title: 'Review', detail: 'Review dari berbagai dimensi' },
    { title: 'Verify', detail: 'Verifikasi temuan penting' },
  ],
}

// Phase 1: Get the diff
phase('Gather')
const diff = await agent(
  'Run `git diff HEAD~1` and return the full diff output. If no diff, try `git diff main`. Return the raw diff text.',
  { label: 'get-diff', phase: 'Gather' }
)

if (!diff || diff.length < 10) {
  return { findings: [], summary: 'No significant changes found.' }
}

log(`Diff size: ${diff.length} chars`)

// Phase 2: Review from multiple angles in parallel
phase('Review')
const FINDINGS_SCHEMA = {
  type: 'object',
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          file: { type: 'string' },
          line: { type: 'number' },
          severity: { type: 'string', enum: ['critical', 'high', 'medium', 'low', 'info'] },
          category: { type: 'string' },
          description: { type: 'string' },
          suggestion: { type: 'string' },
        },
        required: ['file', 'severity', 'description', 'suggestion'],
      },
    },
  },
  required: ['findings'],
}

const dimensions = [
  { key: 'bugs', prompt: 'Review this diff for correctness bugs: logic errors, off-by-one, null reference, race conditions, wrong return values.' },
  { key: 'security', prompt: 'Review this diff for security issues: injection, auth bypass, secrets in code, CORS misconfig, unsafe deserialization.' },
  { key: 'performance', prompt: 'Review this diff for performance issues: N+1 queries, unnecessary loops, missing indexes, large allocations, blocking calls.' },
  { key: 'cleanliness', prompt: 'Review this diff for code quality: dead code, duplicated logic, bad naming, missing error handling, inconsistent style.' },
]

const reviewResults = await parallel(
  dimensions.map(d => () =>
    agent(
      `${d.prompt}\n\nDiff:\n${diff}`,
      { label: `review:${d.key}`, phase: 'Review', schema: FINDINGS_SCHEMA }
    )
  )
)

const allFindings = reviewResults
  .filter(Boolean)
  .flatMap((r, i) => r.findings.map(f => ({ ...f, dimension: dimensions[i].key })))

log(`${allFindings.length} findings across ${dimensions.length} dimensions`)

// Phase 3: Verify critical/high findings
phase('Verify')
const critical = allFindings.filter(f => f.severity === 'critical' || f.severity === 'high')

if (critical.length > 0) {
  const VERDICT_SCHEMA = {
    type: 'object',
    properties: {
      is_real: { type: 'boolean' },
      explanation: { type: 'string' },
    },
    required: ['is_real', 'explanation'],
  }

  const verified = await parallel(
    critical.map(f => () =>
      agent(
        `Verify this finding is real by reading the actual file.\nFile: ${f.file}\nFinding: ${f.description}`,
        { label: `verify:${f.file}`, phase: 'Verify', schema: VERDICT_SCHEMA }
      ).then(v => ({ ...f, verified: v.is_real, verification: v.explanation }))
    )
  )

  const confirmed = verified.filter(v => v.verified)
  const lowPriority = allFindings.filter(f => f.severity !== 'critical' && f.severity !== 'high')

  return { findings: [...confirmed, ...lowPriority], summary: `${confirmed.length} critical/high findings confirmed, ${lowPriority.length} lower priority` }
}

return { findings: allFindings, summary: `${allFindings.length} findings (no critical/high to verify)` }
