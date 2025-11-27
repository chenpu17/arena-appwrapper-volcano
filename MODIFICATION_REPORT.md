# 修改报告 / Modification Report

## Fork 信息 / Fork Information

| 项目 / Item | 值 / Value |
|------------|------------|
| 上游项目 / Upstream | [kubeflow/arena](https://github.com/kubeflow/arena) |
| Fork 基准版本 / Base Version | **v0.15.3** |
| Fork 基准 Commit | `deef4785` (Fix: install script does not work due to CRD files renaming #1391) |
| Fork 日期 / Fork Date | 2024-11 |
| 本项目仓库 / This Repository | [chenpu17/arena-appwrapper-volcano](https://github.com/chenpu17/arena-appwrapper-volcano) |

---

## 代码修改统计 / Code Change Statistics

**总计 / Total: 18 files changed, 3,283 lines inserted**

### 按模块分类 / By Module

| 模块 / Module | 文件数 / Files | 行数 / Lines | 说明 / Description |
|--------------|---------------|-------------|-------------------|
| Helm Chart | 5 | 1,251 | AppWrapper + Volcano Job 模板 |
| API Types | 3 | 320 | 类型定义与 Builder |
| ArgsBuilder | 1 | 394 | 参数构建与验证 |
| Training | 3 | 513 | 训练器与提交逻辑 |
| Commands | 2 | 75 | CLI 命令 |
| K8s Accesser | 1 | 3 | 常量定义 |
| Operators | 2 | 396 | AppWrapper CRD 类型 |
| Client | 1 | 217 | Kubernetes clientset |
| **合计 / Total** | **18** | **3,283** | |

### 详细文件列表 / Detailed File List

```
 charts/appwrapperjob/Chart.yaml                    |    5 +
 charts/appwrapperjob/templates/_helpers.tpl        |   32 +
 charts/appwrapperjob/templates/appwrapper.yaml     | 1049 +
 charts/appwrapperjob/templates/headless-service.yaml |   28 +
 charts/appwrapperjob/values.yaml                   |  137 +
 pkg/apis/arenaclient/training_client.go            |    3 +
 pkg/apis/training/appwrapperjob_builder.go         |  507 +
 pkg/apis/types/submit_appwrapper.go                |  134 +
 pkg/apis/types/training.go                         |    7 +
 pkg/argsbuilder/submit_appwrapper.go               |  394 +
 pkg/commands/training/submit.go                    |    2 +
 pkg/commands/training/submit_appwrapperjob.go      |   73 +
 pkg/k8saccesser/const.go                           |    3 +
 pkg/operators/appwrapper-operator/apis/.../types.go|  179 +
 pkg/operators/appwrapper-operator/client/.../...go |  217 +
 pkg/training/submit_appwrapper.go                  |   64 +
 pkg/training/trainer.go                            |    1 +
 pkg/training/trainer_appwrapper.go                 |  448 +
```

---

## 提交历史 / Commit History

| Commit | 类型 / Type | 描述 / Description |
|--------|------------|-------------------|
| `39cf393c` | feat | 添加 AppWrapper job 支持，集成 Volcano Job |
| `a5dafc86` | fix | 修复 Volcano Job 分布式训练问题，添加文档 |
| `21660228` | docs | 添加 AppWrapper 示例文件 |
| `31233431` | refactor | 将示例移动到 samples 目录 |
| `4fc2ba49` | docs | 重写 README，突出 AppWrapper + Volcano 特性 |

---

## 新增功能 / New Features

### 1. AppWrapper 支持
- **Kueue 集成**: 通过 `--kueue-queue` 指定 LocalQueue 实现资源配额管理
- **故障容错**: 可配置 `--retry-limit`、宽限期、自动恢复机制
- **生命周期管理**: 支持 admission/warmup/failure 等多种宽限期

### 2. Volcano Job 支持
- **Gang 调度**: 通过 `--min-available` 确保分布式训练所有 Pod 同时启动
- **网络拓扑感知**: `--network-topology-mode` 支持 hard/soft 模式
- **分区策略**: `--total-partitions` + `--partition-size` 支持超大规模训练
- **硬件亲和性**: `--ring-controller` 支持华为昇腾等专用 AI 芯片

### 3. 分布式训练增强
- **自动环境变量**: MASTER_ADDR、MASTER_PORT、WORLD_SIZE、RANK
- **Headless Service**: 自动创建用于 Pod DNS 解析的服务
- **双内部作业类型**: 支持 PyTorchJob (`--inner-type pytorch`) 和 Volcano Job (`--inner-type volcano`)

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
RANK 通过 Volcano 的 `volcano.sh/task-index` annotation 获取：

```yaml
- name: RANK
  valueFrom:
    fieldRef:
      fieldPath: metadata.annotations['volcano.sh/task-index']
```

---

## 使用示例 / Usage Examples

详见 [samples/appwrapper/](samples/appwrapper/) 目录：
- `appwrapper.1.yaml` - 完整的 AppWrapper + Volcano Job YAML 示例
- `cli.1.md` - CLI 使用指南与参数对照表

---

## 许可证 / License

本项目继承 Apache License 2.0 许可证。
