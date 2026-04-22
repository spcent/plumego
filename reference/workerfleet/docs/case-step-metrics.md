# Workerfleet Case And Step Metrics Design

本文档定义 workerfleet 面向 pod、worker、exec plan、case 和 step 的完整 Prometheus 指标体系。目标是让 Grafana 能比较全面地展示系统当前状态、pod 实际吞吐、node/pod 维度 case 运行时间分布，以及 case 内关键 step 的耗时分布。

当前代码已经实现基础 worker、pod、case phase、alert 和 runtime 指标。本文档描述的是下一阶段目标指标体系，落地时应继续遵守 workerfleet app-local metrics 边界，不扩展 Plumego 稳定 `metrics` 包。

## 1. 概念口径

- `node`：物理机器，即 Kubernetes node。
- `pod`：Kubernetes pod，一个 pod 内运行一个 worker 容器。
- `worker`：pod 内的业务进程，可以并发运行多个 case。
- `exec_plan_id`：仿真任务计划 ID，一批 case 归属于同一个 exec plan。
- `case`：worker 实际执行的单个仿真单元，当前代码中可继续使用 `task_id` 表示。
- `step`：case 执行过程中的关键步骤，例如准备环境、下载 bundle、数据格式转换、仿真模拟、结果上传、环境清理。

建议固定 step 枚举，避免 Prometheus label 基数失控：

- `prepare_environment`
- `download_bundle`
- `download_dependency_data`
- `validate_input`
- `convert_data_format`
- `run_simulation`
- `collect_outputs`
- `upload_results`
- `cleanup_environment`
- `finalize`

## 2. Label 策略

默认允许 label：

- `namespace`
- `node`
- `pod`
- `worker_state`
- `pod_phase`
- `exec_plan_id`
- `case_type`
- `step`
- `result`
- `error_class`

使用限制：

- `pod` 可以用于状态、吞吐和运行时间分布，因为业务明确要求 pod 维度。
- `exec_plan_id` 可以用于吞吐和耗时分布，但只有在活跃 exec plan 数量受控时默认开启。
- `step` 必须是受控枚举，worker 不能自由上报任意 step 名称。
- `error_class` 必须是低基数分类，例如 `network`、`storage`、`validation`、`simulation`、`timeout`、`unknown`。

禁止作为 Prometheus label：

- `case_id`
- `task_id`
- `worker_id`
- `pod_uid`
- 原始错误信息
- bundle ID
- 数据集 ID

这些高基数字段应进入 MongoDB 明细和 workerfleet 查询 API，而不是进入 Prometheus。

## 3. 当前状态指标

### 3.1 Pod 当前状态

```text
workerfleet_pod_status
```

类型：`gauge`

Labels：

```text
namespace,node,pod,pod_phase
```

`pod_phase`：

- `pending`
- `running`
- `succeeded`
- `failed`
- `deleted`
- `unknown`

语义：

- 当前 pod 所处 phase 的 series 值为 1。
- 同一 pod 的其他 phase series 应为 0 或不存在。

PromQL：

```promql
sum by (node, pod_phase) (workerfleet_pod_status)
```

```promql
sum by (node) (workerfleet_pod_status{pod_phase!="running"})
```

### 3.2 Worker 当前状态

```text
workerfleet_worker_status
```

类型：`gauge`

Labels：

```text
namespace,node,pod,worker_state
```

`worker_state`：

- `running`：进程存活、心跳新鲜、可以接新任务。
- `busy`：进程存活、心跳新鲜、有 active case，但暂不接新任务。
- `abnormal`：进程存活，但有 last error、step stuck 或其他异常信号。
- `lost`：心跳超时，无法确认进程状态。
- `offline`：进程明确不存活，或 pod failed/deleted。
- `unknown`：信号不足。

PromQL：

```promql
sum by (worker_state) (workerfleet_worker_status)
```

```promql
sum by (node, worker_state) (
  workerfleet_worker_status{worker_state!="running",worker_state!="busy"}
)
```

### 3.3 Worker 心跳年龄

```text
workerfleet_worker_heartbeat_age_seconds
```

类型：`gauge`

Labels：

```text
namespace,node,pod
```

语义：当前时间减去 worker 上一次心跳时间。

PromQL：

```promql
max by (node, pod) (workerfleet_worker_heartbeat_age_seconds)
```

告警示例：

```promql
workerfleet_worker_heartbeat_age_seconds > 300
```

### 3.4 Worker 当前并发 case 数

```text
workerfleet_worker_active_cases
```

类型：`gauge`

Labels：

```text
namespace,node,pod,exec_plan_id
```

如果 `exec_plan_id` 基数过高，应拆成两个指标：

```text
workerfleet_worker_active_cases{namespace,node,pod}
workerfleet_exec_plan_active_cases{namespace,exec_plan_id}
```

PromQL：

```promql
sum by (node, pod) (workerfleet_worker_active_cases)
```

```promql
sum by (exec_plan_id) (workerfleet_worker_active_cases)
```

## 4. Pod 维度吞吐指标

### 4.1 Case 开始数量

```text
workerfleet_case_started_total
```

类型：`counter`

Labels：

```text
namespace,node,pod,exec_plan_id,case_type
```

### 4.2 Case 完成数量

```text
workerfleet_case_completed_total
```

类型：`counter`

Labels：

```text
namespace,node,pod,exec_plan_id,case_type,result
```

`result`：

- `succeeded`
- `failed`
- `canceled`
- `timeout`
- `unknown`

每小时成功数：

```promql
sum by (node, pod) (
  increase(workerfleet_case_completed_total{result="succeeded"}[1h])
)
```

每小时失败数：

```promql
sum by (node, pod) (
  increase(workerfleet_case_completed_total{result="failed"}[1h])
)
```

成功率：

```promql
sum by (node, pod) (
  increase(workerfleet_case_completed_total{result="succeeded"}[1h])
)
/
sum by (node, pod) (
  increase(workerfleet_case_completed_total[1h])
)
```

### 4.3 Case 失败分类

```text
workerfleet_case_failed_total
```

类型：`counter`

Labels：

```text
namespace,node,pod,exec_plan_id,case_type,error_class
```

PromQL：

```promql
sum by (node, pod, error_class) (
  increase(workerfleet_case_failed_total[1h])
)
```

## 5. Case 运行时间分布

### 5.1 Case 总耗时

```text
workerfleet_case_duration_seconds
```

类型：`histogram`

Labels：

```text
namespace,node,pod,exec_plan_id,case_type,result
```

语义：一个 case 从开始到结束的总耗时。

推荐 buckets：

```text
30s, 1m, 2m, 5m, 10m, 15m, 30m, 1h, 2h, 4h, 8h
```

Node 维度 p95：

```promql
histogram_quantile(
  0.95,
  sum by (le, node) (
    rate(workerfleet_case_duration_seconds_bucket[5m])
  )
)
```

Pod 维度 p95：

```promql
histogram_quantile(
  0.95,
  sum by (le, node, pod) (
    rate(workerfleet_case_duration_seconds_bucket[5m])
  )
)
```

Exec plan 维度 p95：

```promql
histogram_quantile(
  0.95,
  sum by (le, exec_plan_id) (
    rate(workerfleet_case_duration_seconds_bucket[5m])
  )
)
```

## 6. Step 耗时分布

### 6.1 Step 耗时主指标

```text
workerfleet_case_step_duration_seconds
```

类型：`histogram`

Labels：

```text
namespace,node,pod,exec_plan_id,case_type,step,result
```

语义：case 中某个关键 step 的耗时分布。

推荐 buckets：

```text
1s, 2s, 5s, 10s, 30s, 1m, 2m, 5m, 10m, 30m, 1h, 2h
```

Step p95：

```promql
histogram_quantile(
  0.95,
  sum by (le, step) (
    rate(workerfleet_case_step_duration_seconds_bucket[5m])
  )
)
```

Node + step p95：

```promql
histogram_quantile(
  0.95,
  sum by (le, node, step) (
    rate(workerfleet_case_step_duration_seconds_bucket[5m])
  )
)
```

Pod + step p95：

```promql
histogram_quantile(
  0.95,
  sum by (le, node, pod, step) (
    rate(workerfleet_case_step_duration_seconds_bucket[5m])
  )
)
```

Exec plan + step p95：

```promql
histogram_quantile(
  0.95,
  sum by (le, exec_plan_id, step) (
    rate(workerfleet_case_step_duration_seconds_bucket[5m])
  )
)
```

### 6.2 Step 完成数量

```text
workerfleet_case_step_completed_total
```

类型：`counter`

Labels：

```text
namespace,node,pod,exec_plan_id,case_type,step,result
```

用途：

- 判断 step histogram 是否有样本。
- 查看某个 step 的吞吐和失败集中情况。

PromQL：

```promql
sum by (step, result) (
  increase(workerfleet_case_step_completed_total[1h])
)
```

## 7. 卡住和异常指标

### 7.1 当前 step 卡住 case 数

```text
workerfleet_case_step_stuck_cases
```

类型：`gauge`

Labels：

```text
namespace,node,pod,exec_plan_id,case_type,step,severity
```

`severity`：

- `warning`
- `critical`

PromQL：

```promql
sum by (node, pod, step) (workerfleet_case_step_stuck_cases)
```

### 7.2 当前最老 active step 年龄

```text
workerfleet_case_step_oldest_active_age_seconds
```

类型：`gauge`

Labels：

```text
namespace,node,pod,exec_plan_id,case_type,step
```

语义：某个 label 组合下，当前仍在运行的最老 step 已运行多久。

PromQL：

```promql
max by (node, pod, step) (
  workerfleet_case_step_oldest_active_age_seconds
)
```

告警示例：

```promql
workerfleet_case_step_oldest_active_age_seconds{step="download_bundle"} > 1800
```

## 8. 完整指标清单

状态类：

- `workerfleet_pod_status`
- `workerfleet_worker_status`
- `workerfleet_worker_heartbeat_age_seconds`
- `workerfleet_worker_active_cases`

吞吐类：

- `workerfleet_case_started_total`
- `workerfleet_case_completed_total`
- `workerfleet_case_failed_total`
- `workerfleet_case_step_completed_total`

耗时分布类：

- `workerfleet_case_duration_seconds`
- `workerfleet_case_step_duration_seconds`

异常类：

- `workerfleet_case_step_stuck_cases`
- `workerfleet_case_step_oldest_active_age_seconds`

已有基础指标可继续保留：

- `workerfleet_workers`
- `workerfleet_pods`
- `workerfleet_active_cases`
- `workerfleet_alerts_total`
- `workerfleet_alerts_firing`
- `workerfleet_worker_report_apply_duration_seconds`
- `workerfleet_kube_inventory_sync_duration_seconds`

后续实现时可逐步把旧指标迁移到新语义，避免长期重复。

## 9. Grafana Dashboard 分组

### Fleet Current State

展示当前系统健康度：

- worker state 分布。
- pod phase 分布。
- lost、offline、abnormal worker 数量。
- active cases 总数。
- active cases by node/pod。

核心 PromQL：

```promql
sum by (worker_state) (workerfleet_worker_status)
```

```promql
sum by (node, pod) (workerfleet_worker_active_cases)
```

### Pod Throughput

展示每个 pod 实际产出：

- 每小时成功 case 数。
- 每小时失败 case 数。
- 每小时总完成数。
- 成功率。
- 失败分类。

核心 PromQL：

```promql
sum by (node, pod) (
  increase(workerfleet_case_completed_total{result="succeeded"}[1h])
)
```

```promql
sum by (node, pod) (
  increase(workerfleet_case_completed_total{result="failed"}[1h])
)
```

### Case Duration Distribution

展示业务耗时是否波动：

- case total p50/p95/p99 by node。
- case total p50/p95/p99 by pod。
- case total p95 by exec_plan_id。
- succeeded vs failed 耗时对比。

核心 PromQL：

```promql
histogram_quantile(
  0.95,
  sum by (le, node, pod) (
    rate(workerfleet_case_duration_seconds_bucket[5m])
  )
)
```

### Step Duration Distribution

展示 10 个关键步骤耗时：

- step p50/p95/p99。
- step p95 by node。
- step p95 by pod。
- step p95 by exec_plan_id。
- step result 分布。

核心 PromQL：

```promql
histogram_quantile(
  0.95,
  sum by (le, step) (
    rate(workerfleet_case_step_duration_seconds_bucket[5m])
  )
)
```

```promql
histogram_quantile(
  0.95,
  sum by (le, node, pod, step) (
    rate(workerfleet_case_step_duration_seconds_bucket[5m])
  )
)
```

### Stuck And Abnormal

展示异常定位：

- stuck cases by step。
- oldest active step age。
- abnormal worker by node/pod。
- failure by error_class。

核心 PromQL：

```promql
sum by (node, pod, step) (workerfleet_case_step_stuck_cases)
```

```promql
max by (node, pod, step) (
  workerfleet_case_step_oldest_active_age_seconds
)
```

## 10. Cardinality 风险

最大风险来自：

```text
pod * exec_plan_id * step * result * histogram_bucket
```

8000 个 pod 下，如果活跃 `exec_plan_id` 数量很多，series 数会快速增长。

默认开启建议：

```text
namespace,node,pod,step,result
```

可选开启：

```text
exec_plan_id
```

不进入 Prometheus：

```text
case_id/task_id
worker_id
pod_uid
原始错误信息
bundle_id
dataset_id
```

## 11. 数据生成规则

服务端通过前后 snapshot diff 生成指标：

- case 首次出现：`workerfleet_case_started_total +1`
- case 从 active set 消失：`workerfleet_case_completed_total +1`，记录 `workerfleet_case_duration_seconds`
- step 从 A 切到 B：记录 A 的 `workerfleet_case_step_duration_seconds`
- step failed：`workerfleet_case_step_completed_total{result="failed"} +1`
- step succeeded：`workerfleet_case_step_completed_total{result="succeeded"} +1`
- 当前 step 超过阈值：更新 `workerfleet_case_step_stuck_cases`
- 当前 step 最长运行时间：更新 `workerfleet_case_step_oldest_active_age_seconds`

Worker 上报建议：

```json
{
  "task_id": "case-001",
  "exec_plan_id": "plan-123",
  "task_type": "simulation",
  "phase": "running",
  "current_step": {
    "step": "run_simulation",
    "status": "running",
    "started_at": "2026-04-22T10:00:00Z",
    "updated_at": "2026-04-22T10:05:00Z"
  }
}
```

## 12. 落地优先级

建议按优先级分批落地：

1. 当前状态：`pod_status`、`worker_status`、`heartbeat_age`、`active_cases`
2. Pod 吞吐：`case_started_total`、`case_completed_total`、`case_failed_total`
3. 核心业务耗时：`case_duration_seconds`
4. Step 耗时分布：`case_step_duration_seconds`
5. 异常定位：`case_step_stuck_cases`、`case_step_oldest_active_age_seconds`

这样可以覆盖：

- 当前 pod/worker 是否存活、是否异常、是否失联。
- 每个 pod 每小时成功/失败吞吐。
- 每个 node/pod 的 case 总耗时分布。
- 每个关键 step 的耗时分布。
- 当前业务指标是否出现波动，以及波动集中在哪个 node、pod、step 或 exec plan。

