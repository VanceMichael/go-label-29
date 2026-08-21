# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

财务为一笔货损索赔发起赔付，付款网关返回 unavailable，接口也把这个错误交给了上层；可索赔单已经显示 resolved 并写入完成时间，再点重试只得到状态冲突，实际款项始终没有支付。请先不要修改代码，查清失败请求为何留下不可重试的已结案状态，说明索赔读取、状态保存、外部付款和错误返回的先后关系，以及哪些失败位置会产生这种分裂结果。

## 含 Bug 版本

- 仓库：VanceMichael/go-label-29
- 仓库地址：https://github.com/VanceMichael/go-label-29.git
- parent SHA：0b83477a9cb37103914dd731cca3f2a5dc4d7d67

## 复现步骤

```bash
git clone -- https://github.com/VanceMichael/go-label-29.git bug-repro
cd bug-repro
git checkout --detach 0b83477a9cb37103914dd731cca3f2a5dc4d7d67
go test ./internal/claims -run ^TestSettlementFailureLeavesClaimOpenForRetry$ -count=1
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/claims -run ^TestSettlementFailureLeavesClaimOpenForRetry$ -count=1
--- FAIL: TestSettlementFailureLeavesClaimOpenForRetry (0.00s)
    resolution_test.go:32: failed payout changed claim = {ID:claim-1 TenantID:airline ShipmentID:shipment-1 FiledBy:ops Reason:damaged cargo Status:resolved FiledAt:2026-08-21 06:14:34.571333216 +0000 UTC ResolvedAt:2026-08-21 07:14:34.571333216 +0000 UTC}
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/claims	0.029s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/claims -run ^TestSettlementFailureLeavesClaimOpenForRetry$ -count=1
--- FAIL: TestSettlementFailureLeavesClaimOpenForRetry (0.00s)
    resolution_test.go:32: failed payout changed claim = {ID:claim-1 TenantID:airline ShipmentID:shipment-1 FiledBy:ops Reason:damaged cargo Status:resolved FiledAt:2026-08-21 06:14:56.280473296 +0000 UTC ResolvedAt:2026-08-21 07:14:56.280473296 +0000 UTC}
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/claims	0.001s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

根因结论必须定位理赔协调流程中索赔状态保存早于外部付款且付款失败没有补偿的具体路径，完整解释首次调用为何同时出现网关错误与 resolved 状态、重试为何在状态转换处被拒绝，并区分读取、状态转换、保存和付款各自失败的影响边界；定向命令 go test ./internal/claims -run '^TestSettlementFailureLeavesClaimOpenForRetry$' -count=1 应稳定重现，所有调查基于固定 main SHA，且目标仓库代码、测试和配置零改动。
