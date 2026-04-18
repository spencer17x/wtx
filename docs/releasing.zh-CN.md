# 发布 wtx

## 必需 Secrets

- `NPM_TOKEN`：具备 npm 发布权限
- `HOMEBREW_TAP_TOKEN`：具备 `spencer17x/homebrew-wtx` 仓库推送权限

## 必需外部准备

- `wtx` npm 包名的发布权限
- `spencer17x/homebrew-wtx` tap 仓库

## 日常开发

普通分支 push 和 pull request 只运行 CI，不会发布任何产物。只有 `v*` 标签 push 才会触发正式发布流程。

## 发布前 Dry Run

在推送正式标签之前，可以手动运行 `Release Dry Run` 工作流，对任意分支、标签或提交做一次候选发布验证。

- `ref`：要检出的分支名、已有标签或 commit SHA
- `version`：候选版本号，格式必须是 `vX.Y.Z`

Dry run 会校验版本号格式、运行 `npm test`、运行 `npm run release:check`、检查 `npm pack --dry-run`，并执行 `npm run release:snapshot`。整个过程不会发布任何内容。

## 正式发布

1. 确认 `main` 为绿色状态。
2. 如有需要，先对计划发布的 `ref` 和候选版本运行一次 `Release Dry Run`。
3. 在待发布提交上创建 `vX.Y.Z` 形式的标签。
4. 推送该标签。

```bash
git tag v0.1.0
git push origin v0.1.0
```

该标签会触发：

- GitHub Release 产物发布
- npm 发布
- Homebrew tap 更新

## 重新运行时的行为

GitHub Release 产物是后续发布步骤的前提。

- 如果对应的 GitHub Release 还不存在，工作流会运行 GoReleaser 创建 Release 并上传产物。
- 如果 Release 已存在，且包含预期的平台归档文件和 `checksums.txt`，工作流会跳过 GoReleaser，并继续后续发布步骤。
- 如果 Release 对象已存在，但任何必需产物缺失或为空，工作流会直接失败，不会继续发布 npm，以避免从不完整的 Release 继续向下游分发。
