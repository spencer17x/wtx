# wtx

[English README](./README.md)

`wtx` 是一个基于 Go 的 CLI，用来在 `git worktree add` 之上补齐本地开发环境初始化流程。

创建 worktree 往往只是第一步。实际开发里，团队通常还需要复制本地配置、复用编辑器目录、恢复依赖环境，并避免把构建产物带进新目录。`wtx` 保留 Git 原生 worktree 工作流，同时处理这层初始化问题。

发布说明见：[docs/releasing.md](./docs/releasing.md)

## 项目能力

- 封装 `git worktree add`，但不替代 Git
- 基于已有分支或新分支创建 worktree
- 为项目名和 worktree 目录推导合理默认值
- 通过显式策略处理被忽略的文件和目录：
  - `copy`
  - `symlink`
  - `skip`
  - `setup`
- 自动识别常见项目类型并执行对应初始化命令
- 支持单次覆盖、持久化配置、hooks、dry-run 和批量创建

## 安装

### npm

```bash
npm install -g wtx
```

### Homebrew

```bash
brew tap spencer17x/wtx
brew install wtx
```

### 源码构建

```bash
go build -o bin/wtx ./cmd/wtx
```

运行测试：

```bash
go test ./...
```

## 快速开始

基于已有分支创建一个 worktree：

```bash
./bin/wtx add feature/my-branch
```

创建 worktree 并同时创建新分支：

```bash
./bin/wtx add feature/my-branch --new-branch --base main
```

只预览执行计划，不做任何修改：

```bash
./bin/wtx add feature/my-branch --new-branch --dry-run --non-interactive
```

一次性创建多个 worktree：

```bash
./bin/wtx batch-add feature/one feature/two --new-branch --root ~/worktrees --non-interactive
```

## 命令

### `add`

```bash
./bin/wtx add <branch> [options]
```

创建单个 worktree。

### `batch-add`

```bash
./bin/wtx batch-add <branch> [<branch> ...] [options]
```

按顺序创建多个 worktree，并复用同一组选项。

说明：

- `batch-add` 主要用于非交互场景
- `batch-add` 不支持 `--name` 或 `--dir`
- 如果要统一控制批量创建目录，请使用 `--root`

## 参数

- `--new-branch`
  为 worktree 创建新分支
- `--base <ref>`
  配合 `--new-branch` 指定基线引用
- `--name <name>`
  覆盖默认推导的项目名
- `--dir <path>`
  覆盖目标 worktree 目录
- `--root <path>`
  覆盖默认 worktree 根目录
- `--mode <default|custom|none>`
  选择初始化模式
- `--dry-run`
  只打印完整计划，不创建 worktree，也不执行 setup
- `--non-interactive`
  即使 stdin 是 TTY，也禁用交互提示
- `-y`, `--yes`
  接受默认值并跳过确认
- `-h`, `--help`
  显示帮助

## 初始化模式

- `default`
  使用推导出的默认策略处理已识别的 ignored 路径
- `custom`
  为每个检测到的 ignored 路径单独选择策略
- `none`
  完全跳过复用和 setup

## 复用策略

- `copy`
  将文件或目录复制到新 worktree
- `symlink`
  在新 worktree 中创建指向源路径的软链接
- `skip`
  完全忽略该路径
- `setup`
  不 copy / symlink，而是在新 worktree 中执行初始化命令

## 自动 Setup 检测

`wtx` 会识别常见生态并自动选择初始化命令。

示例：

- Node.js
  - `bun.lock` / `bun.lockb` -> `bun install`
  - `pnpm-lock.yaml` -> `pnpm install`
  - `yarn.lock` -> `yarn install`
  - 只有 `package.json` -> `npm install`
- Python
  - `uv.lock` -> `uv sync`
  - `poetry.lock` -> `poetry install`
  - `Pipfile` / `Pipfile.lock` -> `pipenv install`
  - `requirements.txt` / `pyproject.toml` -> virtualenv + pip 流程
- Go
  - `go.mod` -> `go mod download`
- Rust
  - `Cargo.toml` -> `cargo fetch`
- Java
  - `mvnw` + `pom.xml` -> `./mvnw dependency:resolve`
  - `pom.xml` -> `mvn dependency:resolve`
  - `gradlew` + `build.gradle` -> `./gradlew build`
  - `build.gradle` -> `gradle build`

## 配置

`wtx` 会按顺序读取以下可选 JSON 配置：

- `~/.wtx.json`
- `<repo>/.wtx.json`

项目级配置会覆盖用户级配置。

### 支持字段

```json
{
  "worktreeRoot": "/Users/alex/worktrees",
  "strategyOverrides": {
    ".claude": "symlink",
    ".env": "copy",
    "node_modules": "setup"
  },
  "setupTemplates": {
    "node-pnpm": [
      {
        "id": "node-pnpm-frozen",
        "description": "Install Node.js dependencies with pnpm using the lockfile",
        "command": "pnpm",
        "args": ["install", "--frozen-lockfile"]
      }
    ]
  },
  "hooks": {
    "beforeCreate": [
      {
        "id": "announce-start",
        "description": "Announce worktree creation",
        "command": "echo",
        "args": ["before-create"]
      }
    ],
    "afterCreate": [
      {
        "id": "announce-finish",
        "description": "Announce worktree completion",
        "command": "echo",
        "args": ["after-create"]
      }
    ]
  }
}
```

### `worktreeRoot`

设置默认的 worktree 根目录。

### `strategyOverrides`

覆盖指定 ignored 路径的推导策略。

例如：

- `.env` -> `copy`
- `.claude` -> `symlink`
- `node_modules` -> `setup`

### `setupTemplates`

通过模板 ID 替换自动检测出来的 setup 命令。

例如：

- 将自动检测出的 `node-pnpm` 替换成 `pnpm install --frozen-lockfile`
- 将自动检测出的 `python-uv-sync` 替换成自定义 `uv` 命令

### `hooks`

在 worktree 创建前后执行额外命令。

支持的 hook 阶段：

- `beforeCreate`
- `afterCreate`

Hook 进程会收到以下环境变量：

- `WTX_REPO_ROOT`
- `WTX_WORKTREE_DIRECTORY`
- `WTX_PROJECT_NAME`
- `WTX_BRANCH_NAME`
- `WTX_BRANCH_MODE`

## 安全说明

- `symlink` 会让源目录和新 worktree 指向同一份底层文件
- 构建产物和临时目录通常会被推导为 `skip`
- `node_modules`、`.venv` 这类依赖环境更适合走 `setup`，而不是直接共享
- `--dry-run` 是检查完整计划且不做任何改动的最安全方式

## 示例工作流

使用默认行为创建一个新分支 worktree：

```bash
./bin/wtx add feature/refactor-auth --new-branch
```

先预览再决定是否执行：

```bash
./bin/wtx add feature/refactor-auth --new-branch --dry-run --non-interactive
```

在统一目录下批量创建多个 review worktree：

```bash
./bin/wtx batch-add review/a review/b review/c --new-branch --root ~/worktrees --non-interactive
```

## 项目结构

- [cmd/wtx/main.go](/Users/17admin/projects/wtx/cmd/wtx/main.go:1)
  CLI 入口
- [internal/cli/app.go](/Users/17admin/projects/wtx/internal/cli/app.go:1)
  参数解析、交互流程、执行编排
- [internal/core](/Users/17admin/projects/wtx/internal/core)
  默认值、策略规划、setup 检测、核心类型
- [internal/config/config.go](/Users/17admin/projects/wtx/internal/config/config.go:1)
  配置加载和合并逻辑
- [internal/git](/Users/17admin/projects/wtx/internal/git)
  Git 命令构造与仓库信息读取
- [internal/fsops/apply.go](/Users/17admin/projects/wtx/internal/fsops/apply.go:1)
  copy、symlink 和 setup 执行
