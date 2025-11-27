# Arena AppWrapper CLI 使用指南 / Arena AppWrapper CLI Guide

本文档说明如何使用 Arena CLI 提交与 `appwrapper.1.yaml` 示例等效的任务。

This document explains how to use Arena CLI to submit jobs equivalent to the `appwrapper.1.yaml` example.

---

## 示例 YAML 与 CLI 参数对照表 / YAML to CLI Parameter Mapping

| YAML 字段 / YAML Field | 示例值 / Example Value | CLI 参数 / CLI Parameter |
|------------------------|------------------------|--------------------------|
| `metadata.name` | `sample-wrapper-vcjob` | `--name` |
| `metadata.labels.kueue.x-k8s.io/queue-name` | `team-a-queue` | `--kueue-queue` |
| `spec.components[0].podSets[0].replicas` | `4` | `--replicas` |
| `spec.components[0].template.kind` | `Job` (Volcano) | `--inner-type volcano` |
| `spec.components[0].template.metadata.labels.ring-controller.volcano` | `ascend-1980` | `--ring-controller` |
| `spec.components[0].template.spec.minAvailable` | `4` | `--min-available` |
| `spec.components[0].template.spec.schedulerName` | `volcano` | `--scheduler-name` |
| `spec.components[0].template.spec.networkTopology.mode` | `hard` | `--network-topology-mode` |
| `spec.components[0].template.spec.networkTopology.highestTierAllowed` | `2` | `--highest-tier-allowed` |
| `spec.components[0].template.spec.tasks[0].replicas` | `4` | `--replicas` |
| `spec.components[0].template.spec.tasks[0].name` | `worker1` | `--task-name` |
| `spec.components[0].template.spec.tasks[0].maxRetry` | `10000` | `--max-retry` |
| `spec.components[0].template.spec.tasks[0].partitionPolicy.totalPartitions` | `2` | `--total-partitions` |
| `spec.components[0].template.spec.tasks[0].partitionPolicy.partitionSize` | `2` | `--partition-size` |
| `spec.components[0].template.spec.tasks[0].partitionPolicy.networkTopology.mode` | `hard` | `--partition-topology-mode` |
| `spec.components[0].template.spec.tasks[0].partitionPolicy.networkTopology.highestTierAllowed` | `1` | `--partition-highest-tier` |
| `containers[0].image` | `xx` | `--image` |
| `containers[0].command` | `sleep 3000` | 命令行最后的参数 / Last argument |
| `containers[0].resources.requests.cpu` | `10m` | `--cpu` |
| `containers[0].resources.requests.huawei.com/ascend-1980` | `16` | `--device "huawei.com/ascend-1980=16"` |
| `spec.tolerations` | (见下文) | `--toleration` |

---

## 提交与 appwrapper.1.yaml 等效的任务 / Submit Job Equivalent to appwrapper.1.yaml

```bash
arena submit appwrapperjob \
  --name sample-wrapper-vcjob \
  --image xx \
  --inner-type volcano \
  --replicas 4 \
  --min-available 4 \
  --kueue-queue team-a-queue \
  --scheduler-name volcano \
  --ring-controller ascend-1980 \
  --network-topology-mode hard \
  --highest-tier-allowed 2 \
  --total-partitions 2 \
  --partition-size 2 \
  --partition-topology-mode hard \
  --partition-highest-tier 1 \
  --task-name worker1 \
  --max-retry 10000 \
  --cpu 10m \
  --device "huawei.com/ascend-1980=16" \
  --toleration "node.kubernetes.io/unreachable:NoExecute:Exists:60" \
  --toleration "node.kubernetes.io/not-ready:NoExecute:Exists:60" \
  "sleep 3000"
```

---

## 验证生成的 YAML / Verify Generated YAML

使用 Helm template 命令预览生成的 YAML（无需实际提交）：

Use Helm template command to preview the generated YAML (without actually submitting):

```bash
helm template sample-wrapper-vcjob ./charts/appwrapperjob \
  --set innerJobType=volcano \
  --set replicas=4 \
  --set minAvailable=4 \
  --set kueueQueueName=team-a-queue \
  --set schedulerName=volcano \
  --set ringController=ascend-1980 \
  --set networkTopologyMode=hard \
  --set highestTierAllowed=2 \
  --set totalPartitions=2 \
  --set partitionSize=2 \
  --set partitionNetworkTopologyMode=hard \
  --set partitionHighestTierAllowed=1 \
  --set taskName=worker1 \
  --set maxRetry=10000 \
  --set image=xx \
  --set 'command=sleep 3000'
```

---

## CLI 自动添加的额外配置 / Additional Configurations Added by CLI

CLI 会自动添加以下示例 YAML 中没有但分布式训练所必需的配置：

The CLI automatically adds the following configurations that are not in the example YAML but are required for distributed training:

### 1. Headless Service

用于 Pod DNS 解析，使 Pod 之间可以通过 DNS 名称互相通信。

For Pod DNS resolution, enabling Pods to communicate with each other via DNS names.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: sample-wrapper-vcjob
spec:
  clusterIP: None
  selector:
    release: sample-wrapper-vcjob
```

### 2. Pod subdomain 配置 / Pod subdomain Configuration

```yaml
spec:
  subdomain: sample-wrapper-vcjob
```

### 3. 分布式训练环境变量 / Distributed Training Environment Variables

```yaml
env:
  - name: MASTER_ADDR
    value: "sample-wrapper-vcjob-worker1-0.sample-wrapper-vcjob"
  - name: MASTER_PORT
    value: "23456"
  - name: WORLD_SIZE
    value: "4"
  - name: RANK
    valueFrom:
      fieldRef:
        fieldPath: metadata.annotations['volcano.sh/task-index']
```

### 4. AppWrapper 故障容忍注解 / AppWrapper Fault Tolerance Annotations

```yaml
annotations:
  workload.codeflare.dev.appwrapper/retryLimit: "3"
  workload.codeflare.dev.appwrapper/admissionGracePeriodDuration: 1m
  workload.codeflare.dev.appwrapper/warmupGracePeriodDuration: 5m
  workload.codeflare.dev.appwrapper/failureGracePeriodDuration: 1m
  workload.codeflare.dev.appwrapper/retryPausePeriodDuration: 90s
```

---

## 基本用法示例 / Basic Usage Examples

### 简单的 Volcano Job / Simple Volcano Job

```bash
# 4个副本，使用 GPU / 4 replicas with GPU
arena submit appwrapperjob \
  --name my-training \
  --image pytorch/pytorch:latest \
  --inner-type volcano \
  --replicas 4 \
  --gpus 1 \
  --kueue-queue default-queue \
  "python train.py"
```

### PyTorchJob 模式 / PyTorchJob Mode

```bash
# 使用 PyTorchJob 作为内部类型（默认）
# Use PyTorchJob as inner type (default)
arena submit appwrapperjob \
  --name my-pytorch-job \
  --image pytorch/pytorch:latest \
  --workers 4 \
  --gpus 1 \
  --kueue-queue default-queue \
  "python train.py"
```

---

## 任务管理命令 / Job Management Commands

```bash
# 查看任务状态 / Check job status
arena get my-training --type appwrapperjob

# 查看日志 / View logs
arena logs my-training --type appwrapperjob

# 列出所有 AppWrapper 任务 / List all AppWrapper jobs
arena list --type appwrapperjob

# 删除任务 / Delete job
arena delete my-training --type appwrapperjob
```

---

## 完整参数参考 / Complete Parameter Reference

```bash
arena submit appwrapperjob --help
```

### AppWrapper 特定参数 / AppWrapper Specific Parameters

| 参数 / Parameter | 默认值 / Default | 说明 / Description |
|------------------|------------------|---------------------|
| `--inner-type` | `pytorch` | 内部作业类型 / Inner job type: `pytorch` or `volcano` |
| `--kueue-queue` | - | Kueue LocalQueue 名称 / Kueue LocalQueue name |
| `--retry-limit` | `3` | 最大重试次数 / Maximum retries |
| `--admission-grace-period` | `1m` | Pod 准入等待时间 / Pod admission wait time |
| `--warmup-grace-period` | `5m` | Pod 就绪等待时间 / Pod ready wait time |
| `--failure-grace-period` | `1m` | 故障宽限期 / Failure grace period |
| `--retry-pause-period` | `90s` | 重试间隔 / Retry pause interval |
| `--success-ttl` | - | 成功后自动删除时间 / Auto-delete after success |

### Volcano Job 特定参数 / Volcano Job Specific Parameters

| 参数 / Parameter | 默认值 / Default | 说明 / Description |
|------------------|------------------|---------------------|
| `--replicas` | `1` | 任务副本数 / Task replicas |
| `--min-available` | `replicas` | Gang 调度最小 Pod 数 / Min pods for gang scheduling |
| `--scheduler-name` | `volcano` | 调度器名称 / Scheduler name |
| `--task-name` | `worker` | 任务名称 / Task name |
| `--master-port` | `23456` | 分布式训练通信端口 / Distributed training port |
| `--max-retry` | `10000` | 任务最大重试次数 / Task max retry |
| `--ring-controller` | - | 环形控制器标签 / Ring controller label (e.g., `ascend-1980`) |
| `--network-topology-mode` | - | 网络拓扑模式 / Network topology mode: `hard` or `soft` |
| `--highest-tier-allowed` | `0` | 最高网络拓扑层级 / Highest network tier |
| `--total-partitions` | `0` | 总分区数 / Total partitions |
| `--partition-size` | `0` | 每分区 Pod 数 / Pods per partition |
| `--partition-topology-mode` | - | 分区内网络拓扑模式 / Partition network topology mode |
| `--partition-highest-tier` | `0` | 分区内最高层级 / Partition highest tier |

---

## 前置条件 / Prerequisites

- Kubernetes 集群需安装以下组件 / Kubernetes cluster with:
  - [AppWrapper Operator](https://github.com/project-codeflare/appwrapper) (workload.codeflare.dev/v1beta2)
  - [Kueue](https://kueue.sigs.k8s.io/) (可选 / optional)
  - [Volcano](https://volcano.sh/) >= v1.5 (`--inner-type volcano` 时必需 / required for volcano mode)
  - [Kubeflow Training Operator](https://github.com/kubeflow/training-operator) (`--inner-type pytorch` 时必需 / required for pytorch mode)
