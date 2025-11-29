# 修改报告 / Modification Report

## Fork 信息 / Fork Information

| 项目 / Item | 值 / Value |
|------------|------------|
| 上游项目 / Upstream | [kubeflow/arena](https://github.com/kubeflow/arena) |
| Fork 基准版本 / Base Version | **v0.15.3** |
| Fork 基准 Commit | `deef4785` (2025-11-26, Fix: install script does not work due to CRD files renaming #1391) |
| Fork 日期 / Fork Date | 2025-11-27 |
| 本项目仓库 / This Repository | [chenpu17/arena-appwrapper-volcano](https://github.com/chenpu17/arena-appwrapper-volcano) |

---

## 代码修改统计 / Code Change Statistics

**总计 / Total: 27 files changed, 3,640 lines inserted**

### 按模块分类 / By Module

| 模块 / Module | 文件数 / Files | 行数 / Lines | 说明 / Description |
|--------------|---------------|-------------|-------------------|
| Helm Chart | 5 | 1,257 | AppWrapper + Volcano Job 模板 |
| API Types | 4 | 329 | 类型定义与 Builder |
| ArgsBuilder | 7 | 668 | 参数构建与验证（含 toleration 增强） |
| Training | 4 | 528 | 训练器与提交逻辑 |
| Commands | 2 | 75 | CLI 命令 |
| K8s Accesser | 1 | 3 | 常量定义 |
| Operators | 2 | 396 | AppWrapper CRD 类型 |
| Client | 1 | 217 | Kubernetes clientset |
| Samples | 1 | 80 | 示例文件 |
| **合计 / Total** | **27** | **3,640** | |

### 详细文件列表 / Detailed File List

```
 charts/appwrapperjob/Chart.yaml                    |    5 +
 charts/appwrapperjob/templates/_helpers.tpl        |   32 +
 charts/appwrapperjob/templates/appwrapper.yaml     | 1055 +
 charts/appwrapperjob/templates/headless-service.yaml |   28 +
 charts/appwrapperjob/values.yaml                   |  137 +
 pkg/apis/arenaclient/training_client.go            |    3 +
 pkg/apis/training/appwrapperjob_builder.go         |  507 +
 pkg/apis/types/submit.go                           |    9 +-
 pkg/apis/types/submit_appwrapper.go                |  133 +
 pkg/apis/types/training.go                         |    7 +
 pkg/argsbuilder/evaluatejob.go                     |    5 +-
 pkg/argsbuilder/model.go                           |    5 +-
 pkg/argsbuilder/serving.go                         |    5 +-
 pkg/argsbuilder/submit.go                          |    5 +-
 pkg/argsbuilder/submit_appwrapper.go               |  391 +
 pkg/argsbuilder/update_serving.go                  |    5 +-
 pkg/argsbuilder/util.go                            |  252 +-
 pkg/commands/training/submit.go                    |    2 +
 pkg/commands/training/submit_appwrapperjob.go      |   73 +
 pkg/k8saccesser/const.go                           |    3 +
 pkg/operators/appwrapper-operator/apis/.../types.go|  179 +
 pkg/operators/appwrapper-operator/client/.../...go |  217 +
 pkg/training/submit_appwrapper.go                  |   64 +
 pkg/training/trainer.go                            |    1 +
 pkg/training/trainer_appwrapper.go                 |  453 +
 pkg/training/trainer_volcano.go                    |   10 +
 samples/appwrapper/appwrapper.1.yaml               |   80 +
```

---

## 提交历史 / Commit History

| Commit | 日期 / Date | 类型 / Type | 描述 / Description |
|--------|------------|------------|-------------------|
| `39cf393c` | 2025-11-27 | feat | 添加 AppWrapper job 支持，集成 Volcano Job |
| `a5dafc86` | 2025-11-27 | fix | 修复 Volcano Job 分布式训练问题，添加文档 |
| `21660228` | 2025-11-27 | docs | 添加 AppWrapper 示例文件 |
| `31233431` | 2025-11-27 | refactor | 将示例移动到 samples 目录 |
| `4fc2ba49` | 2025-11-27 | docs | 重写 README，突出 AppWrapper + Volcano 特性 |
| `ce7b5caf` | 2025-11-27 | docs | 添加修改报告 |
| `717275da` | 2025-11-27 | docs | README 关联修改报告 |
| `7f06f867` | 2025-11-27 | fix | 修复 Kubernetes not-found 错误转换 |
| `75f635c0` | 2025-11-27 | fix | 移除重复的 schedulerName 字段 |
| `1b84f589` | 2025-11-27 | docs | 添加安装指南 |
| `59baac6e` | 2025-11-27 | docs | 区分编译环境和运行环境要求 |
| `2df2ad64` | 2025-11-27 | docs | 添加常见问题排查章节 |
| `137329f3` | 2025-11-28 | docs | 添加架构说明：Volcano Job vs PyTorchJob |
| `24188eba` | 2025-11-28 | docs | 改进架构图和时序图 |
| `788783d2` | 2025-11-28 | docs | 更新修改报告：架构图和完整提交历史 |
| `56950611` | 2025-11-28 | docs | 修正 fork 日期（2024→2025），添加详细日期 |
| `f47411f4` | 2025-11-28 | fix | 增强 toleration 解析，支持 tolerationSeconds |
| `8b5264c7` | 2025-11-28 | docs | 添加昇腾 910C 分布式训练完整示例 |
| `620842fa` | 2025-11-29 | docs | 添加存储配置指南（PVC/NFS） |
| `909a6db5` | 2025-11-29 | docs | 更新修改报告统计和提交历史 |
| `0190c7dd` | 2025-11-29 | fix | 启用 Volcano svc 插件修复 Pod DNS 解析 |
| `4c56dabc` | 2025-11-29 | docs | 更新版本引用从 v0.2.0 到 v0.3.0 |
| `4fa32096` | 2025-11-29 | feat | 添加 NNODES 环境变量，修复 WORLD_SIZE 计算 |
| `f6574bb5` | 2025-11-29 | docs | 添加经验证的 Swift SFT 训练示例 |
| `643feb16` | 2025-11-29 | feat | 添加 LOCAL_WORLD_SIZE 环境变量支持 DeepSpeed |
| `e224fe0c` | 2025-11-29 | docs | 更新所有分布式训练示例添加 --nproc-per-node |
| `b99ebd67` | 2025-11-29 | feat | 添加 NODE_RANK 环境变量支持 swift/ms-swift |

---

## 新增功能 / New Features

### 1. AppWrapper 支持
- **Kueue 集成**: 通过 `--kueue-queue` 指定 LocalQueue 实现资源配额管理
- **故障容错**: 可配置 `--retry-limit`、宽限期、自动恢复机制
- **生命周期管理**: 支持 admission/warmup/failure 等多种宽限期

### 2. Volcano Job 支持
- **Gang 调度**: 通过 `--min-available` 确保分布式训练所有 Pod 同时启动
- **网络拓扑感知**: `--network-topology-mode` 支持 hard/soft 模式，用于昇腾超节点亲和
- **分区策略**: `--total-partitions` + `--partition-size` 支持超大规模训练
- **硬件亲和性**: `--ring-controller` 支持华为昇腾等专用 AI 芯片

### 3. 分布式训练增强
- **自动环境变量**: MASTER_ADDR、MASTER_PORT、NNODES、NPROC_PER_NODE、WORLD_SIZE、RANK、NODE_RANK、LOCAL_WORLD_SIZE
- **Swift/ms-swift 兼容**: NODE_RANK 环境变量支持 swift 自动检测分布式配置
- **DeepSpeed 兼容**: LOCAL_WORLD_SIZE 环境变量支持 DeepSpeed 正确识别每节点进程数
- **Headless Service**: 自动创建用于 Pod DNS 解析的服务
- **双内部作业类型**: 支持 PyTorchJob (`--inner-type pytorch`) 和 Volcano Job (`--inner-type volcano`)

---

## 架构设计 / Architecture Design

### 为什么选择 Volcano Job

`networkTopology`（超节点亲和）是 Volcano Job CRD 的专有字段，PyTorchJob 不支持此特性。

### 系统架构

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

### 时序图

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

> **关键点**: AppWrapper 是 Kueue 和 Volcano 之间的桥梁。Volcano 调度器不支持 `suspend` 字段，而 Kueue 需要通过 `suspend` 控制作业准入。AppWrapper 支持 `suspend`，解决了这一兼容性问题。

---

## 关键技术实现 / Key Technical Implementations

### WorkerCount 同步机制
Volcano 模式下，`--replicas` 参数需要同步到 `WorkerCount` 以确保资源统计正确：

```go
// pkg/argsbuilder/submit_appwrapper.go
func (s *SubmitAppWrapperJobArgsBuilder) syncVolcanoWorkerCount() error {
    if s.args.InnerJobType != "volcano" {
        return nil
    }
    s.args.WorkerCount = int(s.args.Replicas)
    s.args.Envs["workers"] = strconv.Itoa(s.args.WorkerCount)
    // ...
}
```

### Headless Service + Subdomain
Volcano Job 的 Pod DNS 解析依赖 Headless Service 和 subdomain 配置：

```yaml
# charts/appwrapperjob/templates/headless-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}
spec:
  clusterIP: None
  selector:
    release: {{ .Release.Name }}
```

```yaml
# Pod spec
spec:
  subdomain: {{ .Release.Name }}
```

### 分布式训练环境变量
RANK 和 NODE_RANK 通过 Volcano 的 `volcano.sh/task-index` annotation 获取：

```yaml
- name: RANK
  valueFrom:
    fieldRef:
      fieldPath: metadata.annotations['volcano.sh/task-index']
- name: NODE_RANK
  valueFrom:
    fieldRef:
      fieldPath: metadata.annotations['volcano.sh/task-index']
- name: LOCAL_WORLD_SIZE
  value: "{{ .Values.nprocPerNode }}"
```

**环境变量完整列表：**

| 环境变量 | 值 | 说明 |
|---------|-----|------|
| `MASTER_ADDR` | `{job}-{task}-0.{job}` | Master 节点 DNS 地址 |
| `MASTER_PORT` | `--master-port` 值 | 分布式通信端口 |
| `NNODES` | `--replicas` 值 | 节点数量 |
| `NPROC_PER_NODE` | `--nproc-per-node` 值 | 每节点进程数 |
| `WORLD_SIZE` | NNODES × NPROC_PER_NODE | 总进程数 |
| `RANK` | volcano.sh/task-index | 当前节点编号 |
| `NODE_RANK` | volcano.sh/task-index | 同 RANK，swift 兼容 |
| `LOCAL_WORLD_SIZE` | `--nproc-per-node` 值 | DeepSpeed 兼容 |

### 错误处理优化
将 Kubernetes "not found" 错误转换为 Arena 标准错误类型：

```go
// pkg/training/trainer_appwrapper.go
if strings.Contains(err.Error(), fmt.Sprintf(`"%v" not found`, name)) {
    return nil, types.ErrTrainingJobNotFound
}
```

---

## 使用示例 / Usage Examples

### 昇腾超节点亲和调度示例

```bash
arena submit appwrapperjob \
  --name network-topology-job \
  --inner-type volcano \
  --scheduler-name volcano \
  --min-available 4 \
  --network-topology-mode hard \
  --highest-tier-allowed 1 \
  --replicas 8 \
  --device "huawei.com/ascend-1980=16" \
  --kueue-queue your-queue \
  --image your-image:latest \
  "python train.py"
```

详见 [samples/appwrapper/](samples/appwrapper/) 目录：
- `appwrapper.1.yaml` - 完整的 AppWrapper + Volcano Job YAML 示例
- `cli.1.md` - CLI 使用指南与参数对照表

---

## 许可证 / License

本项目继承 Apache License 2.0 许可证。
