# GistSync Workflow Refactor Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构同步工作流，支持按文件选择性上传/恢复，并将冲突处理改为显式面板以降低误覆盖风险。

**Architecture:** 保持后端 syncflow 作为业务核心，新增最小必要请求字段支持条目筛选；前端将同步页面拆为流程编排与冲突决策组件，移除 prompt 式冲突交互。通过 typed backend adapter 保持请求/响应一致性。

**Tech Stack:** Go 1.23, Wails v2, Vue 3 + TypeScript + TailwindCSS

---

## Chunk 1: Backend selective sync support

### Task 1: 为上传请求增加条目筛选

**Files:**
- Modify: `internal/syncflow/service.go`
- Test: `internal/syncflow/service_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestService_UploadProfile_WithSelectedItemIDs(t *testing.T) {
    // setup profile with two items and select one
    // expect Uploaded == 1 and snapshot only contains selected item
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/syncflow -run UploadProfile_WithSelectedItemIDs -v`
Expected: FAIL (field/logic missing)

- [ ] **Step 3: Write minimal implementation**

```go
type UploadProfileRequest struct {
    Profile settings.Profile
    MasterPassword string
    SelectedItemIDs []string
}
```

并在 `UploadProfile` 中仅处理选中项（空切片=全部）。

- [ ] **Step 4: Run focused tests**

Run: `go test ./internal/syncflow -run UploadProfile -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/syncflow/service.go internal/syncflow/service_test.go
git commit -m "feat(syncflow): support selective upload by item ids"
```

### Task 2: 为恢复请求增加条目筛选

**Files:**
- Modify: `internal/syncflow/service.go`
- Test: `internal/syncflow/service_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestService_PreviewApplyConflicts_WithSelectedItemIDs(t *testing.T) {}
func TestService_ApplySnapshot_WithSelectedItemIDs(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `go test ./internal/syncflow -run SelectedItemIDs -v`
Expected: FAIL

- [ ] **Step 3: Implement selected item filtering**

在 `ApplySnapshotRequest` 增加 `SelectedItemIDs`，并在冲突预览与应用流程统一过滤。

- [ ] **Step 4: Run syncflow tests**

Run: `go test ./internal/syncflow -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/syncflow/service.go internal/syncflow/service_test.go
git commit -m "feat(syncflow): support selective restore by item ids"
```

## Chunk 2: Wails bridge + frontend API models

### Task 3: 扩展前后端类型与桥接

**Files:**
- Modify: `app.go`
- Modify: `frontend/src/lib/backend.ts`

- [ ] **Step 1: Write TypeScript compile-failing shape usage in UI components**

通过新增 `selectedItemIds` 字段触发类型缺失。

- [ ] **Step 2: Run typecheck to confirm fail**

Run: `npm run -C frontend build`
Expected: FAIL before backend.ts updated

- [ ] **Step 3: Implement bridge updates**

- `UploadProfile` 改为接收请求对象（含 `selectedItemIds`）
- 更新 `ApplySnapshotRequest` 类型

- [ ] **Step 4: Rebuild frontend**

Run: `npm run -C frontend build`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app.go frontend/src/lib/backend.ts
git commit -m "refactor(api): add selected item ids for sync operations"
```

## Chunk 3: Frontend workflow refactor

### Task 4: 拆分页面结构并重组导航

**Files:**
- Modify: `frontend/src/App.vue`
- Create: `frontend/src/components/SyncCenter.vue`
- Create: `frontend/src/components/ProfileManager.vue`
- Modify: `frontend/src/components/SettingsPanel.vue`
- Delete: `frontend/src/components/SyncPanel.vue`

- [ ] **Step 1: Add failing usage snapshot (manual)**
记录当前 UI 问题作为对照。

- [ ] **Step 2: Implement structural split**

`App.vue` 改为三段信息架构：同步中心/配置管理/安全设置。

- [ ] **Step 3: Wire data ownership**

`ProfileManager` 负责配置管理，`SyncCenter` 负责上传/恢复。

- [ ] **Step 4: Build frontend**

Run: `npm run -C frontend build`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/App.vue frontend/src/components/*.vue
git commit -m "refactor(ui): split sync workflow and profile management"
```

### Task 5: 冲突决策面板替代 prompt

**Files:**
- Create: `frontend/src/components/ConflictResolverDialog.vue`
- Modify: `frontend/src/components/SyncCenter.vue`

- [ ] **Step 1: Write failing component integration**

在 `SyncCenter` 使用 `ConflictResolverDialog` props/events。

- [ ] **Step 2: Build to confirm missing component contract**

Run: `npm run -C frontend build`
Expected: FAIL before dialog implementation

- [ ] **Step 3: Implement dialog**

提供逐项覆盖/跳过 + 全选覆盖/全选跳过；默认跳过。

- [ ] **Step 4: Build frontend**

Run: `npm run -C frontend build`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ConflictResolverDialog.vue frontend/src/components/SyncCenter.vue
git commit -m "feat(ui): add explicit conflict resolver dialog"
```

## Chunk 4: Visual cleanup and verification

### Task 6: 统一视觉 token 与反馈样式

**Files:**
- Modify: `frontend/src/style.css`
- Modify: `frontend/src/components/*.vue`

- [ ] **Step 1: Apply spacing/button/status conventions**
- [ ] **Step 2: Ensure mobile/desktop layout usable**
- [ ] **Step 3: Build frontend**

Run: `npm run -C frontend build`
Expected: PASS

- [ ] **Step 4: Run backend tests**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/style.css frontend/src/components/*.vue
git commit -m "style(ui): unify layout and interaction feedback"
```
