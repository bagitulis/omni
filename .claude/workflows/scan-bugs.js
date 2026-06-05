// Workflow: Scan codebase for bugs, verify findings, report
export const meta = {
  name: 'scan-bugs',
  description: 'Scan codebase buat nyari bugs dan verifikasi tiap temuan',
  phases: [
    { title: 'Scan', detail: 'Cari potential bugs di codebase' },
    { title: 'Verify', detail: 'Verifikasi tiap temuan biar gak false positive' },
    { title: 'Report', detail: 'Compile hasil akhir' },
  ],
}

// Phase 1: Scan multiple areas in parallel
phase('Scan')
const BUGS_SCHEMA = {
  type: 'object',
  properties: {
    bugs: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          file: { type: 'string' },
          line: { type: 'number' },
          severity: { type: 'string', enum: ['critical', 'high', 'medium', 'low'] },
          description: { type: 'string' },
          suggested_fix: { type: 'string' },
        },
        required: ['file', 'severity', 'description'],
      },
    },
  },
  required: ['bugs'],
}

const scanResults = await parallel([
  () => agent(
    'Scan backend/ Python files for bugs: unhandled exceptions, SQL injection, missing input validation, resource leaks. Focus on critical and high severity only.',
    { label: 'scan:backend', phase: 'Scan', schema: BUGS_SCHEMA }
  ),
  () => agent(
    'Scan frontend/ files for bugs: XSS vulnerabilities, state management issues, unhandled promise rejections, memory leaks. Focus on critical and high severity only.',
    { label: 'scan:frontend', phase: 'Scan', schema: BUGS_SCHEMA }
  ),
  () => agent(
    'Scan docker-compose files and deployment configs for misconfigurations: exposed ports, missing env vars, wrong network configs.',
    { label: 'scan:infra', phase: 'Scan', schema: BUGS_SCHEMA }
  ),
])

const allBugs = scanResults.filter(Boolean).flatMap(r => r.bugs)
log(`Found ${allBugs.length} potential bugs`)

if (allBugs.length === 0) {
  return { bugs: [], summary: 'No bugs found!' }
}

// Phase 2: Verify each finding with adversarial review
phase('Verify')
const VERDICT_SCHEMA = {
  type: 'object',
  properties: {
    is_real: { type: 'boolean' },
    confidence: { type: 'number' },
    explanation: { type: 'string' },
  },
  required: ['is_real', 'confidence', 'explanation'],
}

const verified = await pipeline(
  allBugs,
  async (bug) => {
    const verdict = await agent(
      `Adversarially verify this bug report. Is it a REAL bug or a false positive?
       File: ${bug.file}
       Bug: ${bug.description}
       Read the actual file and verify.`,
      { label: `verify:${bug.file}`, phase: 'Verify', schema: VERDICT_SCHEMA }
    )
    return { ...bug, verified: verdict.is_real, confidence: verdict.confidence, explanation: verdict.explanation }
  }
)

const confirmedBugs = verified.filter(b => b.verified)
log(`${confirmedBugs.length} bugs confirmed out of ${allBugs.length} potential`)

// Phase 3: Report
phase('Report')
const summary = await agent(
  `Compile a bug report from these confirmed bugs. Group by severity, include file paths and suggested fixes:
   ${JSON.stringify(confirmedBugs, null, 2)}`,
  { label: 'compile-report', phase: 'Report' }
)

return { bugs: confirmedBugs, summary }
