# Arena (AppWrapper + Volcano 增强版)

[![GitHub release](https://img.shields.io/github/v/release/kubeflow/arena)](https://github.com/kubeflow/arena/releases) [![Integration Test](https://github.com/kubeflow/arena/actions/workflows/integration.yaml/badge.svg)](https://github.com/kubeflow/arena/actions/workflows/integration.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/kubeflow/arena)](https://goreportcard.com/report/github.com/kubeflow/arena)

> 本项目是 [Kubeflow Arena](https://github.com/kubeflow/arena) 的增强版分支，新增了对 **AppWrapper** 和 **Volcano Job** 的完整支持，适用于大规模分布式训练场景。
>
> **Fork 基准**: Arena v0.15.3 | **代码修改**: 27 文件, 3,640 行 | 详见 [修改报告](MODIFICATION_REPORT.md)

---

## 中文文档

### 核心特性

本分支在原有 Arena 功能基础上，新增以下核心特性：

#### AppWrapper 支持
- **Kueue 集成**：通过 Kueue LocalQueue 实现资源配额管理和队列调度
- **故障容错**：可配置重试次数、宽限期、自动恢复机制
- **生命周期管理**：支持 admission/warmup/failure 等多种宽限期配置

#### Volcano Job 支持
- **Gang 调度**：确保分布式训练的所有 Pod 同时启动
- **网络拓扑感知**：支持 `hard`/`soft` 模式，优化高性能计算网络延迟
- **分区策略**：支持将 Pod 划分为多个分区执行，适用于超大规模训练
- **硬件亲和性**：支持华为昇腾等专用 AI 芯片的调度标签

#### 分布式训练增强
- **自动环境变量配置**：自动设置 `MASTER_ADDR`、`MASTER_PORT`、`WORLD_SIZE`、`RANK`
- **Headless Service**：自动创建用于 Pod 间 DNS 解析的服务
- **双内部作业类型**：支持 PyTorchJob 和 Volcano Job 作为内部工作负载

### 前置条件

Kubernetes 集群需安装以下组件：

| 组件 | 版本要求 | 说明 |
|------|---------|------|
| [AppWrapper Operator](https://github.com/project-codeflare/appwrapper) | v1beta2 | 必需 |
| [Kueue](https://kueue.sigs.k8s.io/) | - | 可选，用于资源配额管理 |
| [Volcano](https://volcano.sh/) | **>= v1.8** | `--inner-type volcano` 时必需 |
| [Kubeflow Training Operator](https://github.com/kubeflow/training-operator) | - | `--inner-type pytorch` 时必需 |

#### Volcano 配置要求

使用 `--inner-type volcano` 时，需确保：

1. **svc 插件支持**：Volcano Controller 需支持 `svc` 插件（v1.8+ 默认支持）
2. **RBAC 权限**：Volcano Controller 的 ServiceAccount 需具备 Service/Endpoints 的 create/update 权限（标准安装已包含）
3. **Job 名称限制**：Job 名称最长 49 字符（DNS 标签限制）

验证 Volcano 安装：
```bash
# 检查 Volcano 版本
kubectl get deployment -n volcano-system volcano-controller -o jsonpath='{.spec.template.spec.containers[0].image}'

# 检查 Controller 是否正常运行
kubectl get pods -n volcano-system
```

### 快速开始

#### 提交 PyTorchJob（AppWrapper 包装）

```bash
arena submit appwrapperjob \
  --name pytorch-test \
  --image pytorch/pytorch:latest \
  --gpus 1 \
  --workers 2 \
  --kueue-queue default-queue \
  "python train.py"
```

#### 提交 Volcano Job（AppWrapper 包装）

```bash
arena submit appwrapperjob \
  --name volcano-test \
  --image pytorch/pytorch:latest \
  --inner-type volcano \
  --replicas 4 \
  --gpus 1 \
  --kueue-queue default-queue \
  --scheduler-name volcano \
  "python train.py"
```

#### 提交带网络拓扑和分区策略的任务

```bash
arena submit appwrapperjob \
  --name distributed-training \
  --image your-image:latest \
  --inner-type volcano \
  --replicas 4 \
  --min-available 4 \
  --kueue-queue team-a-queue \
  --ring-controller ascend-1980 \
  --network-topology-mode hard \
  --highest-tier-allowed 2 \
  --total-partitions 2 \
  --partition-size 2 \
  --partition-topology-mode hard \
  --partition-highest-tier 1 \
  "python train.py"
```

#### 华为昇腾 910C 超节点分布式训练（完整示例）

以下是一个完整的 4 节点 × 16 卡（共 64 张 NPU）的昇腾 910C 分布式训练任务示例：

```bash
arena submit appwrapperjob \
    --name ascend-910c-training \
    --namespace ai-training \
    --image swr.cn-north-4.myhuaweicloud.com/your-org/training:latest \
    --inner-type volcano \
    --replicas 4 \
    --min-available 4 \
    --device "huawei.com/Ascend910C=16" \
    --kueue-queue team-a-queue \
    --master-port 29500 \
    --ring-controller ascend-910c \
    --network-topology-mode hard \
    --highest-tier-allowed 2 \
    --total-partitions 2 \
    --partition-size 2 \
    --partition-topology-mode hard \
    --partition-highest-tier 1 \
    --cpu 192 \
    --memory 768Gi \
    --share-memory 64Gi \
    --warmup-grace-period 15m \
    --failure-grace-period 5m \
    --toleration "huawei.com/Ascend910C:NoSchedule:Exists" \
    --toleration "node.kubernetes.io/not-ready:NoExecute:Exists:300" \
    --toleration "node.kubernetes.io/unreachable:NoExecute:Exists:300" \
    'torchrun --nnodes=4 --nproc_per_node=16 --master_addr=$MASTER_ADDR --master_port=$MASTER_PORT --node_rank=$RANK train.py'
```

**参数详解：**

| 分类 | 参数 | 值 | 说明 |
|------|------|-----|------|
| **基础** | `--name` | `ascend-910c-training` | 任务名称 |
| | `--namespace` | `ai-training` | Kubernetes 命名空间 |
| | `--image` | `swr.cn-north-4...` | 训练容器镜像 |
| | `--inner-type` | `volcano` | 内部使用 Volcano Job |
| **分布式** | `--replicas` | `4` | 4 个训练节点 |
| | `--min-available` | `4` | Gang 调度要求全部就绪 |
| | `--master-port` | `29500` | PyTorch 分布式通信端口 |
| **NPU** | `--device` | `huawei.com/Ascend910C=16` | 每节点 16 张 910C 卡 |
| **Kueue** | `--kueue-queue` | `team-a-queue` | 资源配额队列 |
| **拓扑亲和** | `--ring-controller` | `ascend-910c` | 匹配 910C 节点标签 |
| | `--network-topology-mode` | `hard` | 强制拓扑约束 |
| | `--highest-tier-allowed` | `2` | 允许超节点内调度 |
| **分区策略** | `--total-partitions` | `2` | 分为 2 个分区 |
| | `--partition-size` | `2` | 每分区 2 个 Pod |
| | `--partition-topology-mode` | `hard` | 分区内强制拓扑 |
| | `--partition-highest-tier` | `1` | 分区内同机柜 |
| **资源** | `--cpu` | `192` | 每 Pod 192 核 |
| | `--memory` | `768Gi` | 每 Pod 768GB 内存 |
| | `--share-memory` | `64Gi` | HCCL 共享内存 |
| **容错** | `--warmup-grace-period` | `15m` | 预热宽限 15 分钟 |
| | `--failure-grace-period` | `5m` | 失败宽限 5 分钟 |
| **容忍** | `--toleration` | `huawei.com/Ascend910C:...` | 允许调度到 910C 节点 |
| | `--toleration` | `node.kubernetes.io/not-ready:...` | 节点故障容忍 300 秒 |

**网络拓扑层级说明：**

| 层级 (Tier) | 含义 | 网络延迟 |
|------------|------|---------|
| 0 | 同一节点内 | 最低 |
| 1 | 同一机柜内 | 低 |
| 2 | 同一超节点内 | 中等 |
| 3+ | 跨超节点 | 较高 |

> **提示**：不同集群的设备资源名称可能不同，请通过 `kubectl describe node <node-name>` 确认实际资源名称（如 `huawei.com/Ascend910C`、`huawei.com/Ascend910B` 等）。

### 参数说明

#### AppWrapper 参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--inner-type` | `pytorch` | 内部作业类型：`pytorch` 或 `volcano` |
| `--kueue-queue` | - | Kueue LocalQueue 名称 |
| `--retry-limit` | `3` | 最大重试次数 |
| `--admission-grace-period` | `1m` | Pod 准入等待时间 |
| `--warmup-grace-period` | `5m` | Pod 就绪等待时间 |
| `--failure-grace-period` | `1m` | 故障宽限期 |
| `--retry-pause-period` | `90s` | 重试间隔 |
| `--success-ttl` | - | 成功后自动删除时间 |

#### Volcano Job 参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--replicas` | `1` | 任务副本数 |
| `--min-available` | `replicas` | Gang 调度最小 Pod 数 |
| `--scheduler-name` | `volcano` | 调度器名称 |
| `--task-name` | `worker` | 任务名称 |
| `--master-port` | `23456` | 分布式训练通信端口 |
| `--max-retry` | `10000` | 任务最大重试次数 |
| `--use-svc-plugin` | `true` | 使用 Volcano svc 插件（需 >= 1.8），设为 `false` 回退到手动 Headless Service |
| `--ring-controller` | - | 环形控制器标签（如 `ascend-1980`） |
| `--network-topology-mode` | - | 网络拓扑模式：`hard` 或 `soft` |
| `--highest-tier-allowed` | `0` | 最高网络拓扑层级 |
| `--total-partitions` | `0` | 总分区数 |
| `--partition-size` | `0` | 每分区 Pod 数 |
| `--partition-topology-mode` | - | 分区内网络拓扑模式 |
| `--partition-highest-tier` | `0` | 分区内最高层级 |

### 存储配置

训练任务通常需要挂载外部存储来访问代码、数据集和保存模型。Arena 支持两种存储挂载方式：

#### 存储参数

| 参数 | 格式 | 说明 |
|------|------|------|
| `--data` | `<pvc-name>:<容器路径>` | 挂载 PVC（推荐） |
| `--data-dir` | `<主机路径>` | 挂载 HostPath（路径相同） |

#### 方式 1：使用 PVC（推荐）

**步骤 1**：创建 NFS 存储资源

```yaml
# nfs-storage.yaml
---
# PersistentVolume - 定义 NFS 服务器
apiVersion: v1
kind: PersistentVolume
metadata:
  name: training-data-pv
  labels:
    storage: training-data
spec:
  capacity:
    storage: 1Ti
  accessModes:
    - ReadWriteMany           # NFS 支持多节点读写
  persistentVolumeReclaimPolicy: Retain
  nfs:
    server: 192.168.1.100     # NFS 服务器地址
    path: /exports/training   # NFS 导出路径
  mountOptions:
    - nfsvers=4.1
    - hard
    - timeo=600

---
# PersistentVolumeClaim - 绑定 PV
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: training-data
  namespace: ai-training
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Ti
  selector:
    matchLabels:
      storage: training-data
```

```bash
# 创建存储资源
kubectl apply -f nfs-storage.yaml

# 验证 PVC 状态
kubectl get pvc -n ai-training
# NAME            STATUS   VOLUME              CAPACITY   ACCESS MODES
# training-data   Bound    training-data-pv    1Ti        RWX
```

**步骤 2**：提交任务时挂载 PVC

```bash
arena submit appwrapperjob \
    --name my-training \
    --namespace ai-training \
    --data training-data:/data \
    --data model-store:/models \
    --image your-image:latest \
    'python /data/train.py --output /models'
```

#### 方式 2：使用 HostPath

如果 NFS 已挂载到所有节点的相同路径：

```bash
arena submit appwrapperjob \
    --name my-training \
    --data-dir /mnt/nfs/datasets \
    --data-dir /mnt/nfs/models \
    --image your-image:latest \
    'python /mnt/nfs/datasets/train.py'
```

> **注意**：`--data-dir` 的主机路径和容器路径相同。

#### 完整示例（带存储挂载）

```bash
arena submit appwrapperjob \
    --name ascend-910c-training \
    --namespace ai-training \
    --image swr.cn-north-4.myhuaweicloud.com/your-org/training:latest \
    --inner-type volcano \
    --replicas 4 \
    --min-available 4 \
    --device "huawei.com/Ascend910C=16" \
    --kueue-queue team-a-queue \
    --data training-data:/data \
    --data checkpoints:/checkpoints \
    --cpu 192 \
    --memory 768Gi \
    --share-memory 64Gi \
    'torchrun --nnodes=4 --nproc_per_node=16 \
        --master_addr=$MASTER_ADDR --master_port=$MASTER_PORT --node_rank=$RANK \
        /data/train.py --checkpoint-dir /checkpoints'
```

#### 存储设计说明

Kubernetes 使用 PV/PVC 实现存储抽象：

| 概念 | 创建者 | 职责 |
|------|--------|------|
| **PV** (PersistentVolume) | 运维 | 定义存储后端（NFS 地址、容量、访问模式） |
| **PVC** (PersistentVolumeClaim) | 开发者 | 声明存储需求（多大、什么访问模式） |
| **StorageClass** | 运维 | 动态供给模板（可选，自动创建 PV） |

**优势**：应用只引用 PVC 名称，不关心底层是 NFS/Ceph/云盘，便于跨环境迁移。

### 任务管理

```bash
# 列出所有 AppWrapper 任务
arena list --type appwrapperjob

# 查看任务详情
arena get my-job --type appwrapperjob

# 查看日志
arena logs my-job --type appwrapperjob

# 删除任务
arena delete my-job --type appwrapperjob
```

### 示例文件

更多示例请参考 [samples/appwrapper](samples/appwrapper/) 目录：
- `appwrapper.1.yaml` - 完整的 AppWrapper + Volcano Job 示例
- `cli.1.md` - CLI 使用指南和参数对照表

### 架构说明：为什么选择 Volcano Job 而不是 PyTorchJob

本项目支持两种内部作业类型：`--inner-type pytorch` 和 `--inner-type volcano`。对于需要**昇腾超节点亲和性**的场景，必须使用 Volcano Job。

#### 技术原因

`networkTopology`（超节点亲和）是 Volcano Job CRD 的专有字段：

```yaml
# Volcano Job - 支持 networkTopology ✅
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
spec:
  networkTopology:           # Volcano Job 级别字段
    mode: hard
    highestTierAllowed: 1    # 亲和到 HyperNode tier=1
  minAvailable: 4            # Gang 调度
```

PyTorchJob (Kubeflow) 没有 `networkTopology` 字段，即使使用 Volcano 调度器也无法获得此能力。

#### 功能对比

| 特性 | `--inner-type volcano` | `--inner-type pytorch` |
|-----|------------------------|------------------------|
| 超节点亲和 (networkTopology) | ✅ 原生支持 | ❌ 不支持 |
| Gang 调度 (minAvailable) | ✅ 原生支持 | ⚠️ 需要额外 PodGroup |
| 分区策略 (partitionPolicy) | ✅ 原生支持 | ❌ 不支持 |
| 分布式环境变量 | ✅ 模板自动配置 | ✅ Kubeflow 自动配置 |
| Kueue 配额管理 | ✅ 通过 AppWrapper | ✅ 通过 AppWrapper |

#### 推荐选择

| 场景 | 推荐 |
|-----|------|
| 昇腾 NPU + 超节点亲和 | `--inner-type volcano` |
| 需要 Gang 调度 | `--inner-type volcano` |
| 已有 Kubeflow 生态，无特殊调度需求 | `--inner-type pytorch` |

#### 架构图

```
Arena (CLI)
    │
    │ 创建
    ↓
AppWrapper ←───────── Kueue (监听 + 控制 suspend)
    │                    │
    │                    │ 1. 检查 LocalQueue 配额
    │                    │ 2. 配额足够 → suspend: false
    │                    │ 3. 配额不足 → suspend: true (排队)
    │
    │ suspend=false 时创建
    ↓
Volcano Job
    │ 支持 networkTopology, minAvailable, partitionPolicy
    │
    │ 调度
    ↓
Volcano Scheduler
    │ 超节点亲和调度
    ↓
Pods (昇腾 NPU 节点)
```

#### 时序图

```
User        Arena       K8s API      Kueue        AppWrapper    Volcano Job
  │           │            │           │              │              │
  │──submit──→│            │           │              │              │
  │           │──create───→│           │              │              │
  │           │            │──watch───→│              │              │
  │           │            │           │──check quota─│              │
  │           │            │           │──unsuspend──→│              │
  │           │            │           │              │──create─────→│
  │           │            │           │              │              │──schedule→Pods
```

> **注意**：AppWrapper 是 Kueue 和 Volcano 之间的桥梁。Volcano 调度器不支持 `suspend` 字段，而 Kueue 需要通过 `suspend` 控制作业准入。AppWrapper 支持 `suspend`，解决了这一兼容性问题。

---

## English Documentation

### Core Features

This fork extends the original Arena with the following features:

#### AppWrapper Support
- **Kueue Integration**: Resource quota management via Kueue LocalQueues
- **Fault Tolerance**: Configurable retry limits, grace periods, and automatic recovery
- **Lifecycle Management**: Support for admission/warmup/failure grace periods

#### Volcano Job Support
- **Gang Scheduling**: Ensure all Pods start simultaneously for distributed training
- **Network Topology Awareness**: Support `hard`/`soft` modes for HPC network optimization
- **Partition Policies**: Partition Pods for ultra-large scale training
- **Hardware Affinity**: Support for Huawei Ascend and other AI accelerator scheduling labels

#### Distributed Training Enhancements
- **Automatic Environment Variables**: Auto-configure `MASTER_ADDR`, `MASTER_PORT`, `WORLD_SIZE`, `RANK`
- **Headless Service**: Auto-create service for Pod DNS resolution
- **Dual Inner Job Types**: Support both PyTorchJob and Volcano Job as inner workloads

### Prerequisites

Kubernetes cluster with the following components:

| Component | Version | Notes |
|-----------|---------|-------|
| [AppWrapper Operator](https://github.com/project-codeflare/appwrapper) | v1beta2 | Required |
| [Kueue](https://kueue.sigs.k8s.io/) | - | Optional, for resource quota |
| [Volcano](https://volcano.sh/) | >= v1.5 | Required for `--inner-type volcano` |
| [Kubeflow Training Operator](https://github.com/kubeflow/training-operator) | - | Required for `--inner-type pytorch` |

### Quick Start

#### Submit PyTorchJob (wrapped in AppWrapper)

```bash
arena submit appwrapperjob \
  --name pytorch-test \
  --image pytorch/pytorch:latest \
  --gpus 1 \
  --workers 2 \
  --kueue-queue default-queue \
  "python train.py"
```

#### Submit Volcano Job (wrapped in AppWrapper)

```bash
arena submit appwrapperjob \
  --name volcano-test \
  --image pytorch/pytorch:latest \
  --inner-type volcano \
  --replicas 4 \
  --gpus 1 \
  --kueue-queue default-queue \
  --scheduler-name volcano \
  "python train.py"
```

#### Submit with Network Topology and Partition Policy

```bash
arena submit appwrapperjob \
  --name distributed-training \
  --image your-image:latest \
  --inner-type volcano \
  --replicas 4 \
  --min-available 4 \
  --kueue-queue team-a-queue \
  --ring-controller ascend-1980 \
  --network-topology-mode hard \
  --highest-tier-allowed 2 \
  --total-partitions 2 \
  --partition-size 2 \
  --partition-topology-mode hard \
  --partition-highest-tier 1 \
  "python train.py"
```

#### Huawei Ascend 910C HyperNode Distributed Training (Complete Example)

Here is a complete example for 4 nodes × 16 NPUs (64 NPUs total) distributed training on Ascend 910C:

```bash
arena submit appwrapperjob \
    --name ascend-910c-training \
    --namespace ai-training \
    --image swr.cn-north-4.myhuaweicloud.com/your-org/training:latest \
    --inner-type volcano \
    --replicas 4 \
    --min-available 4 \
    --device "huawei.com/Ascend910C=16" \
    --kueue-queue team-a-queue \
    --master-port 29500 \
    --ring-controller ascend-910c \
    --network-topology-mode hard \
    --highest-tier-allowed 2 \
    --total-partitions 2 \
    --partition-size 2 \
    --partition-topology-mode hard \
    --partition-highest-tier 1 \
    --cpu 192 \
    --memory 768Gi \
    --share-memory 64Gi \
    --warmup-grace-period 15m \
    --failure-grace-period 5m \
    --toleration "huawei.com/Ascend910C:NoSchedule:Exists" \
    --toleration "node.kubernetes.io/not-ready:NoExecute:Exists:300" \
    --toleration "node.kubernetes.io/unreachable:NoExecute:Exists:300" \
    'torchrun --nnodes=4 --nproc_per_node=16 --master_addr=$MASTER_ADDR --master_port=$MASTER_PORT --node_rank=$RANK train.py'
```

**Parameter Details:**

| Category | Parameter | Value | Description |
|----------|-----------|-------|-------------|
| **Basic** | `--name` | `ascend-910c-training` | Job name |
| | `--namespace` | `ai-training` | Kubernetes namespace |
| | `--image` | `swr.cn-north-4...` | Training container image |
| | `--inner-type` | `volcano` | Use Volcano Job internally |
| **Distributed** | `--replicas` | `4` | 4 training nodes |
| | `--min-available` | `4` | Gang scheduling requires all ready |
| | `--master-port` | `29500` | PyTorch distributed communication port |
| **NPU** | `--device` | `huawei.com/Ascend910C=16` | 16 × 910C NPUs per node |
| **Kueue** | `--kueue-queue` | `team-a-queue` | Resource quota queue |
| **Topology** | `--ring-controller` | `ascend-910c` | Match 910C node labels |
| | `--network-topology-mode` | `hard` | Enforce topology constraints |
| | `--highest-tier-allowed` | `2` | Allow scheduling within HyperNode |
| **Partition** | `--total-partitions` | `2` | Split into 2 partitions |
| | `--partition-size` | `2` | 2 Pods per partition |
| | `--partition-topology-mode` | `hard` | Enforce partition topology |
| | `--partition-highest-tier` | `1` | Same rack within partition |
| **Resources** | `--cpu` | `192` | 192 cores per Pod |
| | `--memory` | `768Gi` | 768GB memory per Pod |
| | `--share-memory` | `64Gi` | Shared memory for HCCL |
| **Fault Tolerance** | `--warmup-grace-period` | `15m` | 15 min warmup grace |
| | `--failure-grace-period` | `5m` | 5 min failure grace |
| **Tolerations** | `--toleration` | `huawei.com/Ascend910C:...` | Allow scheduling on 910C nodes |
| | `--toleration` | `node.kubernetes.io/not-ready:...` | Tolerate node failures for 300s |

**Network Topology Tiers:**

| Tier | Meaning | Network Latency |
|------|---------|-----------------|
| 0 | Within same node | Lowest |
| 1 | Within same rack | Low |
| 2 | Within same HyperNode | Medium |
| 3+ | Across HyperNodes | Higher |

> **Tip**: Device resource names may vary across clusters. Use `kubectl describe node <node-name>` to verify actual resource names (e.g., `huawei.com/Ascend910C`, `huawei.com/Ascend910B`, etc.).

### Parameters

#### AppWrapper Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `--inner-type` | `pytorch` | Inner job type: `pytorch` or `volcano` |
| `--kueue-queue` | - | Kueue LocalQueue name |
| `--retry-limit` | `3` | Maximum retries |
| `--admission-grace-period` | `1m` | Pod admission wait time |
| `--warmup-grace-period` | `5m` | Pod ready wait time |
| `--failure-grace-period` | `1m` | Failure grace period |
| `--retry-pause-period` | `90s` | Retry pause interval |
| `--success-ttl` | - | Auto-delete after success |

#### Volcano Job Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `--replicas` | `1` | Task replicas |
| `--min-available` | `replicas` | Min pods for gang scheduling |
| `--scheduler-name` | `volcano` | Scheduler name |
| `--task-name` | `worker` | Task name |
| `--master-port` | `23456` | Distributed training port |
| `--max-retry` | `10000` | Task max retry |
| `--ring-controller` | - | Ring controller label (e.g., `ascend-1980`) |
| `--network-topology-mode` | - | Network topology: `hard` or `soft` |
| `--highest-tier-allowed` | `0` | Highest network tier |
| `--total-partitions` | `0` | Total partitions |
| `--partition-size` | `0` | Pods per partition |
| `--partition-topology-mode` | - | Partition network topology mode |
| `--partition-highest-tier` | `0` | Partition highest tier |

### Job Management

```bash
# List all AppWrapper jobs
arena list --type appwrapperjob

# Get job details
arena get my-job --type appwrapperjob

# View logs
arena logs my-job --type appwrapperjob

# Delete job
arena delete my-job --type appwrapperjob
```

### Examples

For more examples, see [samples/appwrapper](samples/appwrapper/):
- `appwrapper.1.yaml` - Complete AppWrapper + Volcano Job example
- `cli.1.md` - CLI guide with YAML-to-CLI parameter mapping

### Architecture: Why Volcano Job over PyTorchJob

This project supports two inner job types: `--inner-type pytorch` and `--inner-type volcano`. For scenarios requiring **Ascend HyperNode affinity**, Volcano Job is required.

#### Technical Reason

`networkTopology` (HyperNode affinity) is a Volcano Job CRD-specific field:

```yaml
# Volcano Job - supports networkTopology ✅
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
spec:
  networkTopology:           # Volcano Job level field
    mode: hard
    highestTierAllowed: 1    # Affinity to HyperNode tier=1
  minAvailable: 4            # Gang scheduling
```

PyTorchJob (Kubeflow) does not have the `networkTopology` field. Even with Volcano scheduler, this capability is unavailable.

#### Feature Comparison

| Feature | `--inner-type volcano` | `--inner-type pytorch` |
|---------|------------------------|------------------------|
| HyperNode affinity (networkTopology) | ✅ Native support | ❌ Not supported |
| Gang scheduling (minAvailable) | ✅ Native support | ⚠️ Requires PodGroup |
| Partition policy (partitionPolicy) | ✅ Native support | ❌ Not supported |
| Distributed env vars | ✅ Template configured | ✅ Kubeflow configured |
| Kueue quota management | ✅ Via AppWrapper | ✅ Via AppWrapper |

#### Recommendation

| Scenario | Recommended |
|----------|-------------|
| Ascend NPU + HyperNode affinity | `--inner-type volcano` |
| Gang scheduling required | `--inner-type volcano` |
| Existing Kubeflow ecosystem, no special scheduling | `--inner-type pytorch` |

#### Architecture Diagram

```
Arena (CLI)
    │
    │ creates
    ↓
AppWrapper ←───────── Kueue (watches + controls suspend)
    │                    │
    │                    │ 1. Check LocalQueue quota
    │                    │ 2. Quota available → suspend: false
    │                    │ 3. Quota insufficient → suspend: true (queued)
    │
    │ creates when suspend=false
    ↓
Volcano Job
    │ Supports networkTopology, minAvailable, partitionPolicy
    │
    │ schedules
    ↓
Volcano Scheduler
    │ HyperNode affinity scheduling
    ↓
Pods (Ascend NPU Nodes)
```

#### Sequence Diagram

```
User        Arena       K8s API      Kueue        AppWrapper    Volcano Job
  │           │            │           │              │              │
  │──submit──→│            │           │              │              │
  │           │──create───→│           │              │              │
  │           │            │──watch───→│              │              │
  │           │            │           │──check quota─│              │
  │           │            │           │──unsuspend──→│              │
  │           │            │           │              │──create─────→│
  │           │            │           │              │              │──schedule→Pods
```

> **Note**: AppWrapper bridges Kueue and Volcano. Volcano scheduler doesn't support the `suspend` field, while Kueue requires `suspend` for admission control. AppWrapper supports `suspend`, resolving this compatibility issue.

---

## 原有功能 / Original Features

Arena 原有功能保持不变，支持以下训练类型：

The original Arena features remain unchanged, supporting:

- TensorFlow Job
- PyTorch Job
- MPI Job
- Horovod Job
- Spark Job
- ET Job (Elastic Training)

详细文档请参考 / For detailed documentation: [Arena Documentation](https://arena-docs.readthedocs.io/en/latest)

---

## 开发指南 / Development

```bash
# 编译 / Build
make arena

# 编译后的二进制文件位于 / Binary located at
./bin/arena-v*-$(go env GOOS)-$(go env GOARCH)
```

---

## 安装指南 / Installation

### 方式一：使用安装脚本（推荐）/ Method 1: Using Installer (Recommended)

```bash
# 1. 编译并创建安装包 / Build and create installer
make arena-installer

# 2. 解压安装包 / Extract installer
tar -xzf arena-installer-*.tar.gz
cd arena-installer-*

# 3. 运行安装脚本 / Run installer
# 仅安装 CLI 和 charts / Install CLI and charts only
./install.sh --only-binary

# 或完整安装（包含 operators）/ Or full installation (with operators)
# ./install.sh
```

安装脚本自动完成 / Installer automatically handles:
- 安装 arena 二进制到 `/usr/local/bin/arena`
- 安装 `arena-kubectl` 和 `arena-helm`
- 安装 Helm charts 到 `/charts`（root）或 `~/charts`（非 root）

### 方式二：手动安装 / Method 2: Manual Installation

```bash
# 1. 编译 / Build
make arena

# 2. 安装二进制 / Install binary
sudo cp bin/arena-* /usr/local/bin/arena

# 3. 安装 charts / Install charts
# root 用户 / For root user:
sudo cp -r charts /charts
# 非 root 用户 / For non-root user:
cp -r charts ~/charts

# 4. 创建 arena-kubectl 链接 / Create arena-kubectl symlink
sudo ln -s $(which kubectl) /usr/local/bin/arena-kubectl

# 5. 验证安装 / Verify installation
arena version
arena-kubectl version --client
ls ~/charts/appwrapperjob  # 或 /charts/appwrapperjob

# 6. 验证 charts 版本 / Verify charts version (重要!)
# 确保输出包含 "v0.2.0" 和 "svc plugin"
grep -E "version:|svc plugin" ~/charts/appwrapperjob/Chart.yaml
# 或
grep "v0.2.0" /charts/appwrapperjob/templates/appwrapper.yaml
```

### 环境要求 / Requirements

**编译环境 / Build Requirements:**
| 依赖 / Dependency | 版本 / Version | 说明 / Notes |
|------------------|----------------|--------------|
| Go | >= 1.22 | 仅编译需要 / Only for building |

**运行环境 / Runtime Requirements:**
| 依赖 / Dependency | 版本 / Version | 说明 / Notes |
|------------------|----------------|--------------|
| kubectl | - | 集群访问 / Cluster access |
| Helm | >= 3.0 | Chart 渲染（已内置）/ Chart rendering (bundled) |

> **注意 / Note**: 使用预编译安装包时，不需要安装 Go。Helm 已包含在安装包中。
>
> When using pre-built installer, Go is NOT required. Helm is bundled in the installer.

---

## 常见问题 / Troubleshooting

### 0. Charts 版本过旧 / Charts version outdated

**问题 / Problem:** 安装后 charts 仍是旧版本，缺少 svc plugin 支持。

**诊断 / Diagnose:**
```bash
# 检查 charts 版本
grep "version:" ~/charts/appwrapperjob/Chart.yaml   # 或 /charts/
# 如果显示 0.1.0，说明是旧版本
# 如果显示 0.2.0，说明是新版本（包含 svc plugin）

# 或检查模板文件
head -1 ~/charts/appwrapperjob/templates/appwrapper.yaml
# 新版本应显示: {{- /* Chart: appwrapperjob v0.2.0 - Volcano svc plugin support */ -}}
```

**原因 / Cause:** 安装包是基于旧代码构建的，charts 没有更新。

**解决方案 / Solution:**
```bash
# 方案 1：重新打包安装（推荐）
git pull origin master
make arena-installer
# 分发新的安装包到目标机器

# 方案 2：手动更新 charts（快速修复）
# 在开发机器上：
scp -r charts user@target-machine:/tmp/
# 在目标机器上：
sudo rm -rf /charts && sudo cp -r /tmp/charts /charts
# 或
rm -rf ~/charts && cp -r /tmp/charts ~/charts
```

### 1. Charts 目录未找到 / Charts directory not found

**错误 / Error:**
```
ERRO[0000] failed to load chart /charts/appwrapperjob: no such file or directory
```

**解决方案 / Solution:**
```bash
# 复制 charts 到正确位置 / Copy charts to correct location
# root 用户 / For root:
sudo cp -r charts /charts
# 非 root 用户 / For non-root:
cp -r charts ~/charts
```

### 2. arena-kubectl 未找到 / arena-kubectl not found

**错误 / Error:**
```
ERRO[0000] exec: "arena-kubectl": executable file not found in $PATH
```

**解决方案 / Solution:**
```bash
# 创建符号链接 / Create symlink
sudo ln -s $(which kubectl) /usr/local/bin/arena-kubectl
```

### 3. 资源已存在 / Resource already exists

**错误 / Error:**
```
ERRO[0000] Service "xxx" in namespace "default" exists and cannot be imported into the current release
```

**解决方案 / Solution:**
```bash
# 删除已存在的任务后重新提交 / Delete existing job and resubmit
arena delete <job-name> --type appwrapperjob

# 或手动删除 / Or delete manually
kubectl delete service <job-name> -n <namespace>
kubectl delete appwrapper <job-name> -n <namespace>
```

### 4. MLflow 服务未找到（可忽略）/ MLflow service not found (can be ignored)

**错误 / Error:**
```
ERRO[0000] failed to create proxied model client: no mlflow service in any namespace found
```

**说明 / Note:** 这是可选的 MLflow 集成功能，不影响 AppWrapper 任务提交。可以安全忽略。

This is an optional MLflow integration feature and does not affect AppWrapper job submission. Can be safely ignored.

### 5. DNS 解析失败 / DNS resolution failed

**错误 / Error:**
```
[c10d] The IPv6 network addresses of (xxx-worker-0.xxx, 29500) cannot be retrieved (gai error: -2 - Name or service not known)
```

**原因 / Cause:** Pod 的 hostname 未正确设置，导致分布式训练无法解析 MASTER_ADDR。

**诊断 / Diagnose:**
```bash
# 检查 Pod hostname/subdomain
kubectl get pod -l release=<job-name> -o jsonpath='{range .items[*]}{.metadata.name}: hostname={.spec.hostname}, subdomain={.spec.subdomain}{"\n"}{end}'
```

**解决方案 / Solution:**

**方案 A**：使用 Volcano >= 1.8（推荐）
```bash
arena delete <job-name> --type appwrapperjob
arena submit appwrapperjob --inner-type volcano ...
```

**方案 B**：回退到手动 Headless Service（适用于旧版 Volcano）
```bash
arena delete <job-name> --type appwrapperjob
arena submit appwrapperjob --inner-type volcano --use-svc-plugin=false ...
```

---

## 许可证 / License

Apache License 2.0
