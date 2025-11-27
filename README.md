# Arena (AppWrapper + Volcano 增强版)

[![GitHub release](https://img.shields.io/github/v/release/kubeflow/arena)](https://github.com/kubeflow/arena/releases) [![Integration Test](https://github.com/kubeflow/arena/actions/workflows/integration.yaml/badge.svg)](https://github.com/kubeflow/arena/actions/workflows/integration.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/kubeflow/arena)](https://goreportcard.com/report/github.com/kubeflow/arena)

> 本项目是 [Kubeflow Arena](https://github.com/kubeflow/arena) 的增强版分支，新增了对 **AppWrapper** 和 **Volcano Job** 的完整支持，适用于大规模分布式训练场景。

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
| [Volcano](https://volcano.sh/) | >= v1.5 | `--inner-type volcano` 时必需 |
| [Kubeflow Training Operator](https://github.com/kubeflow/training-operator) | - | `--inner-type pytorch` 时必需 |

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
| `--ring-controller` | - | 环形控制器标签（如 `ascend-1980`） |
| `--network-topology-mode` | - | 网络拓扑模式：`hard` 或 `soft` |
| `--highest-tier-allowed` | `0` | 最高网络拓扑层级 |
| `--total-partitions` | `0` | 总分区数 |
| `--partition-size` | `0` | 每分区 Pod 数 |
| `--partition-topology-mode` | - | 分区内网络拓扑模式 |
| `--partition-highest-tier` | `0` | 分区内最高层级 |

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

## 许可证 / License

Apache License 2.0
