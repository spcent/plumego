import { useState, useEffect } from 'react'
import {
  getHealth,
  getStats,
  runDoctor,
  ALL_CHECKS,
  type HealthResult,
  type StatsResult,
  type DoctorResult,
  type CheckResult,
} from '../api/system'

const STATUS_COLORS: Record<string, string> = {
  ok: 'bg-green-100 text-green-700 border-green-300',
  warning: 'bg-yellow-100 text-yellow-700 border-yellow-300',
  error: 'bg-red-100 text-red-700 border-red-300',
  disabled: 'bg-muted text-muted-foreground border-border',
}

function StatusBadge({ label, status }: { label: string; status: string }) {
  return (
    <div className="flex items-center justify-between p-3 rounded-lg border border-border bg-background">
      <span className="text-sm font-medium">{label}</span>
      <span className={`text-xs px-2 py-0.5 rounded border ${STATUS_COLORS[status] ?? STATUS_COLORS.disabled}`}>
        {status}
      </span>
    </div>
  )
}

function StatBox({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex flex-col items-center p-4 rounded-lg border border-border bg-background">
      <span className="text-2xl font-semibold font-mono text-foreground">{value.toLocaleString()}</span>
      <span className="text-xs text-muted-foreground mt-1">{label}</span>
    </div>
  )
}

export default function SystemPage() {
  const [health, setHealth] = useState<HealthResult | null>(null)
  const [stats, setStats] = useState<StatsResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Doctor state
  const [selectedChecks, setSelectedChecks] = useState<string[]>(ALL_CHECKS)
  const [doctorResult, setDoctorResult] = useState<DoctorResult | null>(null)
  const [runningDoctor, setRunningDoctor] = useState(false)
  const [expandedCheck, setExpandedCheck] = useState<string | null>(null)

  useEffect(() => {
    loadAll()
  }, [])

  async function loadAll() {
    setLoading(true)
    setError('')
    try {
      const [h, s] = await Promise.all([getHealth(), getStats()])
      setHealth(h)
      setStats(s)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load system data')
    } finally {
      setLoading(false)
    }
  }

  function toggleCheck(name: string) {
    setSelectedChecks(prev =>
      prev.includes(name) ? prev.filter(c => c !== name) : [...prev, name]
    )
  }

  async function handleRunDoctor() {
    setRunningDoctor(true)
    setError('')
    setDoctorResult(null)
    try {
      const result = await runDoctor(selectedChecks.length > 0 ? selectedChecks : undefined)
      setDoctorResult(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Doctor check failed')
    } finally {
      setRunningDoctor(false)
    }
  }

  function CheckCard({ check }: { check: CheckResult }) {
    const isExpanded = expandedCheck === check.name
    const hasIssues = check.items && check.items.length > 0

    return (
      <div className="border border-border rounded-lg overflow-hidden">
        <div
          className="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-muted/20"
          onClick={() => setExpandedCheck(isExpanded ? null : check.name)}
        >
          <span className={`text-xs px-2 py-0.5 rounded border ${STATUS_COLORS[check.status] ?? STATUS_COLORS.disabled}`}>
            {check.status}
          </span>
          <span className="text-sm font-medium flex-1">{check.name}</span>
          <span className="text-xs text-muted-foreground">
            {check.failed} / {check.total} failed
          </span>
          {hasIssues && (
            <span className="text-xs text-muted-foreground">
              {isExpanded ? '▼' : '▶'}
            </span>
          )}
        </div>

        {isExpanded && hasIssues && (
          <div className="px-4 pb-3 border-t border-border bg-muted/10">
            <div className="mt-2 space-y-1">
              {check.items.map((item, idx) => (
                <div key={idx} className="text-xs text-muted-foreground border-l-2 border-border pl-3 py-1">
                  <div className="font-medium text-foreground">{item.issue}</div>
                  {item.document_id && (
                    <div className="font-mono text-xs opacity-70">doc: {item.document_id}</div>
                  )}
                  {item.entity_id && (
                    <div className="font-mono text-xs opacity-70">entity: {item.entity_id}</div>
                  )}
                  {item.detail && (
                    <div className="opacity-70">{item.detail}</div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="h-full overflow-y-auto bg-background">
      <div className="max-w-3xl mx-auto px-6 py-6 space-y-6">
        <div>
          <h2 className="text-base font-semibold text-foreground">System Observability</h2>
          <p className="text-xs text-muted-foreground mt-0.5">
            Health checks, statistics, and consistency diagnostics
          </p>
        </div>

        {error && (
          <div className="text-sm text-destructive border border-destructive/30 rounded px-3 py-2">
            {error}
          </div>
        )}

        {loading && (
          <div className="text-sm text-muted-foreground">Loading…</div>
        )}

        {!loading && (
          <>
            {/* Health Section */}
            <section>
              <h3 className="text-sm font-semibold mb-3">Health Status</h3>
              {health && (
                <div className="space-y-2">
                  <StatusBadge label="Overall" status={health.status} />
                  <div className="grid grid-cols-2 gap-2">
                    <StatusBadge label="Database" status={health.database} />
                    <StatusBadge label="Storage" status={health.storage} />
                    <StatusBadge label="Search" status={health.search} />
                    <StatusBadge label="AI" status={health.ai} />
                  </div>
                </div>
              )}
            </section>

            {/* Stats Section */}
            <section>
              <h3 className="text-sm font-semibold mb-3">Statistics</h3>
              {stats && (
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                  <StatBox label="Documents" value={stats.documents} />
                  <StatBox label="Versions" value={stats.versions} />
                  <StatBox label="Collections" value={stats.collections} />
                  <StatBox label="Tags" value={stats.tags} />
                  <StatBox label="Import Jobs" value={stats.import_jobs} />
                  <StatBox label="Indexed" value={stats.indexed_documents} />
                  <StatBox label="Pending Index" value={stats.pending_indexes} />
                  <StatBox label="Failed Index" value={stats.failed_indexes} />
                  <StatBox label="AI Tasks" value={stats.ai_tasks} />
                  <StatBox label="Prompts" value={stats.prompts} />
                </div>
              )}
            </section>

            {/* Doctor Section */}
            <section>
              <h3 className="text-sm font-semibold mb-3">Doctor Checks</h3>
              <div className="border border-border rounded-lg p-4 space-y-3 bg-background">
                <p className="text-xs text-muted-foreground">
                  Select checks to run (read-only, no mutations):
                </p>
                <div className="grid grid-cols-2 gap-2">
                  {ALL_CHECKS.map(name => (
                    <label key={name} className="flex items-center gap-2 text-xs cursor-pointer hover:bg-muted/30 rounded px-2 py-1">
                      <input
                        type="checkbox"
                        checked={selectedChecks.includes(name)}
                        onChange={() => toggleCheck(name)}
                        className="accent-primary"
                      />
                      <span className="font-mono">{name}</span>
                    </label>
                  ))}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={handleRunDoctor}
                    disabled={runningDoctor || selectedChecks.length === 0}
                    className="px-4 py-2 text-sm font-medium bg-primary text-primary-foreground rounded hover:opacity-90 disabled:opacity-50"
                  >
                    {runningDoctor ? 'Running…' : 'Run Doctor'}
                  </button>
                  <button
                    onClick={() => setSelectedChecks(ALL_CHECKS)}
                    className="px-4 py-2 text-sm border border-border rounded hover:bg-accent"
                  >
                    Select All
                  </button>
                  <button
                    onClick={() => setSelectedChecks([])}
                    className="px-4 py-2 text-sm border border-border rounded hover:bg-accent"
                  >
                    Clear
                  </button>
                </div>
              </div>

              {doctorResult && (
                <div className="mt-4 space-y-2">
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-medium">Overall Status:</span>
                    <span className={`text-xs px-2 py-0.5 rounded border ${STATUS_COLORS[doctorResult.status] ?? STATUS_COLORS.disabled}`}>
                      {doctorResult.status}
                    </span>
                  </div>
                  <div className="space-y-2">
                    {doctorResult.checks.map(check => (
                      <CheckCard key={check.name} check={check} />
                    ))}
                  </div>
                </div>
              )}
            </section>

            <div className="flex justify-end">
              <button
                onClick={loadAll}
                className="px-4 py-2 text-sm border border-border rounded hover:bg-accent"
              >
                Refresh
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
