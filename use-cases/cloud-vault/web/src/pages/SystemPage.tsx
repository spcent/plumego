import { useEffect, useState } from 'react'
import {
  ALL_CHECKS,
  getHealth,
  getStats,
  runDoctor,
  type CheckResult,
  type DoctorResult,
  type HealthResult,
  type StatsResult,
} from '../api/system'
import BackupPanel from '../components/BackupPanel'
import { Badge, Button, MetricCard, PageFrame, Panel, SkeletonRows, StatusBanner, cn } from '../components/ui'

function statusTone(status: string): 'neutral' | 'success' | 'warning' | 'danger' {
  if (status === 'ok') return 'success'
  if (status === 'warning') return 'warning'
  if (status === 'error') return 'danger'
  return 'neutral'
}

function HealthRow({ label, status }: { label: string; status: string }) {
  return (
    <div className="flex items-center justify-between gap-3 border-b border-border px-3 py-2.5 last:border-b-0">
      <span className="text-sm font-medium text-foreground">{label}</span>
      <Badge tone={statusTone(status)}>{status}</Badge>
    </div>
  )
}

export default function SystemPage() {
  const [health, setHealth] = useState<HealthResult | null>(null)
  const [stats, setStats] = useState<StatsResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

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
    setSelectedChecks(prev => prev.includes(name) ? prev.filter(c => c !== name) : [...prev, name])
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

  return (
    <PageFrame
      title="System Observability"
      description="Health, backups, statistics, and read-only consistency diagnostics."
      action={<Button variant="secondary" size="sm" icon="refresh" onClick={loadAll}>Refresh</Button>}
      width="wide"
    >
      {error && <StatusBanner tone="danger">{error}</StatusBanner>}
      {loading && <Panel><SkeletonRows count={6} /></Panel>}

      {!loading && (
        <>
          <div className="grid gap-5 lg:grid-cols-[0.9fr_1.3fr]">
            <Panel
              title="Health Status"
              description="Service availability by subsystem."
              action={health && <Badge tone={statusTone(health.status)}>{health.status}</Badge>}
            >
              {health && (
                <div className="divide-y divide-border">
                  <HealthRow label="Overall" status={health.status} />
                  <HealthRow label="Database" status={health.database} />
                  <HealthRow label="Storage" status={health.storage} />
                  <HealthRow label="Search" status={health.search} />
                  <HealthRow label="AI" status={health.ai} />
                </div>
              )}
            </Panel>

            <Panel title="Statistics" description="Current vault and automation inventory.">
              {stats && (
                <div className="grid gap-3 p-4 sm:grid-cols-2 lg:grid-cols-5">
                  <MetricCard label="Documents" value={stats.documents.toLocaleString()} />
                  <MetricCard label="Versions" value={stats.versions.toLocaleString()} />
                  <MetricCard label="Collections" value={stats.collections.toLocaleString()} />
                  <MetricCard label="Tags" value={stats.tags.toLocaleString()} />
                  <MetricCard label="Imports" value={stats.import_jobs.toLocaleString()} />
                  <MetricCard label="Indexed" value={stats.indexed_documents.toLocaleString()} tone="success" />
                  <MetricCard label="Pending Index" value={stats.pending_indexes.toLocaleString()} tone="warning" />
                  <MetricCard label="Failed Index" value={stats.failed_indexes.toLocaleString()} tone="danger" />
                  <MetricCard label="AI Tasks" value={stats.ai_tasks.toLocaleString()} tone="accent" />
                  <MetricCard label="Prompts" value={stats.prompts.toLocaleString()} />
                </div>
              )}
            </Panel>
          </div>

          <BackupPanel />

          <Panel title="Doctor Checks" description="Run read-only consistency checks against selected domains.">
            <div className="space-y-4 p-4">
              <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
                {ALL_CHECKS.map(name => (
                  <label
                    key={name}
                    className={cn(
                      'flex cursor-pointer items-center gap-2 rounded-md border px-2.5 py-2 text-xs transition-colors hover:bg-accent',
                      selectedChecks.includes(name) ? 'border-primary/30 bg-primary/10' : 'border-border bg-background/45',
                    )}
                  >
                    <input
                      type="checkbox"
                      checked={selectedChecks.includes(name)}
                      onChange={() => toggleCheck(name)}
                      className="accent-primary"
                    />
                    <span className="truncate font-mono">{name}</span>
                  </label>
                ))}
              </div>
              <div className="flex flex-wrap gap-2">
                <Button onClick={handleRunDoctor} disabled={runningDoctor || selectedChecks.length === 0} variant="primary" icon="check">
                  {runningDoctor ? 'Running...' : 'Run Doctor'}
                </Button>
                <Button onClick={() => setSelectedChecks(ALL_CHECKS)} variant="secondary">Select All</Button>
                <Button onClick={() => setSelectedChecks([])} variant="ghost">Clear</Button>
              </div>

              {doctorResult && (
                <div className="space-y-3 border-t border-border pt-4">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-foreground">Overall status</span>
                    <Badge tone={statusTone(doctorResult.status)}>{doctorResult.status}</Badge>
                  </div>
                  <div className="space-y-2">
                    {doctorResult.checks.map(check => (
                      <CheckCard
                        key={check.name}
                        check={check}
                        expanded={expandedCheck === check.name}
                        onToggle={() => setExpandedCheck(expandedCheck === check.name ? null : check.name)}
                      />
                    ))}
                  </div>
                </div>
              )}
            </div>
          </Panel>
        </>
      )}
    </PageFrame>
  )
}

function CheckCard({ check, expanded, onToggle }: { check: CheckResult; expanded: boolean; onToggle: () => void }) {
  const hasIssues = check.items && check.items.length > 0

  return (
    <div className="overflow-hidden rounded-lg border border-border bg-background/45">
      <button type="button" className="flex w-full items-center gap-3 px-3 py-2.5 text-left hover:bg-accent/70" onClick={onToggle}>
        <Badge tone={statusTone(check.status)}>{check.status}</Badge>
        <span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground">{check.name}</span>
        <span className="text-xs text-muted-foreground">{check.failed} / {check.total} failed</span>
        {hasIssues && <span className="text-xs text-muted-foreground">{expanded ? 'Hide' : 'Show'}</span>}
      </button>

      {expanded && hasIssues && (
        <div className="space-y-2 border-t border-border bg-surface p-3">
          {check.items.map((item, idx) => (
            <div key={idx} className="border-l-2 border-border pl-3 text-xs leading-5 text-muted-foreground">
              <div className="font-medium text-foreground">{item.issue}</div>
              {item.document_id && <div className="font-mono">doc: {item.document_id}</div>}
              {item.entity_id && <div className="font-mono">entity: {item.entity_id}</div>}
              {item.detail && <div>{item.detail}</div>}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
