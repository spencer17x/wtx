# 发布 wtx

## 必需 Secrets

- `NPM_TOKEN`：具备 npm 发布权限

## 必需外部准备

- `@spencer17x/wtx` npm 包名的发布权限

## 日常开发

Pull Request 与 `main` push 只运行 CI，不会发布任何产物；功能分支 push
由对应 Pull Request 验证，避免重复运行。`v*` 标签 push 会同时运行 CI 并触发
正式发布流程。正式发布会在发布前重新执行完整仓库检查，因此发布安全不依赖
并行 CI 任务先完成。

## 发布前 Dry Run

在推送正式标签之前，可以手动运行 `Release Dry Run` 工作流，对任意分支、标签或提交做一次候选发布验证。

- `ref`：要检出的分支名、已有标签或 commit SHA
- `version`：候选版本号，格式必须是 `vX.Y.Z`

Dry run 会校验版本号格式、运行 `npm run check`、检查 `npm pack --dry-run`，并执行 `npm run release:snapshot`。整个过程不会发布任何内容。

## 正式发布

1. 运行 `npm run check`，并确认 `main` 为绿色状态。
2. 如有需要，先对计划发布的 `ref` 和候选版本运行一次 `Release Dry Run`。
3. 在待发布提交上创建 `vX.Y.Z` 形式的标签。
4. 推送该标签。

```bash
git tag v0.1.0
git push origin v0.1.0
```

该标签会触发：

- GitHub Release 产物发布
- `@spencer17x/wtx` 的 npm 发布

工作流会在 GoReleaser 运行期间保持 git 工作区干净；只有 GitHub Release 步骤完成后，才会从标签派生 npm 包版本并执行 npm 发布。

## 重新运行时的行为

重新运行时，会先检查 GitHub Release 产物是否完整，再决定是否继续 npm 发布。

- 如果对应的 GitHub Release 还不存在，工作流会运行 GoReleaser 创建 Release 并上传产物。
- 如果 Release 已存在，且包含预期的平台归档文件和 `checksums.txt`，工作流会跳过 GoReleaser，并继续后续发布步骤。
- 如果 Release 对象已存在，但任何必需产物缺失或为空，工作流会直接失败，不会继续发布 npm，以避免从不完整的 Release 继续向下游分发。
