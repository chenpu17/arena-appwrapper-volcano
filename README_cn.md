# Arena

[![Integration Test](https://github.com/kubeflow/arena/actions/workflows/integration.yaml/badge.svg)](https://github.com/kubeflow/arena/actions/workflows/integration.yaml)[![Go Report Card](https://goreportcard.com/badge/github.com/kubeflow/arena)](https://goreportcard.com/report/github.com/kubeflow/arena)

## 概述

Arena 是一个命令行工具，可供数据科学家轻而易举地运行和监控机器学习训练作业，并便捷地检查结果。目前，它支持单机/分布式深度学习模型训练。在实现层面，它基于 Kubernetes、helm 和 Kubeflow。但数据科学家可能对于 kubernetes 知之甚少。

与此同时，用户需要 GPU 资源和节点管理。Arena 还提供了 `top` 命令，用于检查 Kubernetes 集群内的可用 GPU 资源。

简而言之，Arena 的目标是让数据科学家感觉自己就像是在一台机器上工作，而实际上还可以享受到 GPU 集群的强大力量。

## 设置

您可以按照 [安装指南](https://arena-docs.readthedocs.io/en/latest/installation) 执行操作

## 用户指南

Arena 是一种命令行界面，支持轻而易举地运行和监控机器学习训练作业，并便捷地检查结果。目前，它支持独立/分布式训练。

- [1.使用 git 上的源代码运行训练作业](https://arena-docs.readthedocs.io/en/latest/training/tfjob/standalone/)
- [2.使用 tensorboard 运行训练作业](https://arena-docs.readthedocs.io/en/latest/training/tfjob/tensorboard/)
- [3.运行分布式训练作业](https://arena-docs.readthedocs.io/en/latest/training/tfjob/distributed/)
- [4.使用外部数据运行分布式训练作业](https://arena-docs.readthedocs.io/en/latest/training/tfjob/dataset/)
- [5.运行基于 MPI 的分布式训练作业](https://arena-docs.readthedocs.io/en/latest/training/mpijob/distributed/)
- [6.使用群调度器运行分布式 TensorFlow 训练作业](https://arena-docs.readthedocs.io/en/latest/training/etjob/elastictraining-tensorflow2-mnist/)
- [7.运行 TensorFlow Serving](https://arena-docs.readthedocs.io/en/latest/serving/tfserving/serving/)

## AppWrapper 作业支持（新特性）

本分支新增了对 **AppWrapper** 作业的支持，集成了 **Kueue** 资源配额管理和 **Volcano Job** 作为内部工作负载类型。AppWrapper（来自 CodeFlare 项目）提供了高级作业生命周期管理、故障容错和通过 Kueue 实现的资源配额管理。

### 特性

- **Kueue 集成**：将作业提交到 Kueue LocalQueue 进行资源配额管理
- **双内部作业类型**：支持 PyTorchJob 和 Volcano Job 作为内部工作负载
- **故障容错**：可配置的重试次数、宽限期和自动恢复
- **Volcano Job 支持**：Gang 调度、网络拓扑感知和分区策略
- **分布式训练**：自动设置环境变量（MASTER_ADDR、RANK、WORLD_SIZE）

### 前置条件

- Kubernetes 集群需安装以下组件：
  - [AppWrapper Operator](https://github.com/project-codeflare/appwrapper)（workload.codeflare.dev/v1beta2）
  - [Kueue](https://kueue.sigs.k8s.io/)（可选，用于资源配额管理）
  - [Volcano](https://volcano.sh/) >= v1.5（使用 `--inner-type volcano` 时必需）
  - [Kubeflow Training Operator](https://github.com/kubeflow/training-operator)（使用 `--inner-type pytorch` 时必需）

### 快速开始

```bash
# 提交包装在 AppWrapper 中的 PyTorchJob
arena submit appwrapperjob \
  --name pytorch-test \
  --image pytorch/pytorch:latest \
  --gpus 1 \
  --workers 2 \
  --kueue-queue default-queue \
  "python train.py"

# 提交包装在 AppWrapper 中的 Volcano Job
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

### AppWrapper 特定选项

| 选项 | 默认值 | 描述 |
|------|--------|------|
| `--inner-type` | pytorch | 内部作业类型：`pytorch` 或 `volcano` |
| `--kueue-queue` | - | Kueue LocalQueue 名称，用于资源配额 |
| `--retry-limit` | 3 | 标记为失败前的最大重试次数 |
| `--admission-grace-period` | 1m | 等待 Pod 准入的时间 |
| `--warmup-grace-period` | 5m | 等待 Pod 就绪的时间 |
| `--failure-grace-period` | 1m | 将故障视为永久性之前的宽限期 |
| `--retry-pause-period` | 90s | 重试之间的暂停时间 |
| `--success-ttl` | - | 成功后自动删除的时间 |

### Volcano Job 特定选项

| 选项 | 默认值 | 描述 |
|------|--------|------|
| `--replicas` | 1 | 任务副本数 |
| `--min-available` | replicas | Gang 调度所需的最小 Pod 数 |
| `--scheduler-name` | volcano | 调度器名称 |
| `--task-name` | worker | Volcano Job 中的任务名称 |
| `--master-port` | 23456 | 分布式训练通信端口 |
| `--network-topology-mode` | - | 网络拓扑模式：`hard` 或 `soft` |
| `--highest-tier-allowed` | 0 | 网络拓扑层级限制 |
| `--total-partitions` | 0 | 分布式任务的总分区数 |
| `--partition-size` | 0 | 每个分区的 Pod 数量 |

### 管理 AppWrapper 作业

```bash
# 列出所有 AppWrapper 作业
arena list --type appwrapperjob

# 获取作业详情
arena get pytorch-test --type appwrapperjob

# 查看日志
arena logs pytorch-test --type appwrapperjob

# 删除作业
arena delete pytorch-test --type appwrapperjob
```

## 演示

[![arena demo](demo.jpg)](http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/50210690772.mp4)

## 开发

先决条件：

- Go >= 1.8

```shell
mkdir -p $(go env GOPATH)/src/github.com/kubeflow
cd $(go env GOPATH)/src/github.com/kubeflow
git clone https://github.com/kubeflow/arena.git
cd arena
make
```

`arena` 二进制文件位于 `arena/bin` 目录下。您可能希望将目录添加到 `$PATH`。

## 命令行文档

请参阅 [arena.md](docs/cli/arena.md)

## 路线图

请参阅[路线图](ROADMAP.md)
