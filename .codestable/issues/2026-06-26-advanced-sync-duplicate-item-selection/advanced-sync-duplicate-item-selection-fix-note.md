---
doc_type: issue-fix
issue: 2026-06-26-advanced-sync-duplicate-item-selection
path: fast-track
fix_date: 2026-06-26
tags: [sync, upload, restore, item-id]
---

# 高级同步同批文件无法单独勾选修复记录

## 1. 问题描述

高级同步页里，同一批加入配置的多个文件会共用同一个条目 ID，导致上传列表和恢复列表里的勾选状态被绑在一起，只能整批选中或整批取消。

## 2. 根因

`internal/profileutil/profileutil.go` 之前用 `time.Now().UnixNano()` 生成条目 ID。Windows 下同一轮快速多次调用可能拿到相同时间戳，于是同批文件写入了重复 ID。前端复选框和后端选择集合都以该 ID 为身份键，重复后就会被当成同一个条目。

## 3. 修复方案

将条目身份改为“基于路径计算的稳定 ID”，并在本地配置与云端 manifest 加载/保存时统一做归一化：

- 新增文件直接生成稳定 ID，不再依赖时间戳。
- 加载本地配置时自动修复历史重复 ID。
- 读取/保存云端 manifest 时按同一规则重写 profile item 和 snapshot item 的 ID，兼容旧快照里的重复 ID。
- 增加针对历史重复 ID 场景的回归测试。

## 4. 改动文件清单

- `internal/profileutil/profileutil.go`
- `internal/appsvc/default_profile_manager.go`
- `internal/appsvc/service_test.go`
- `internal/syncflow/manifest.go`
- `internal/syncflow/service_test.go`

## 5. 验证结果

- `go test ./internal/appsvc ./internal/syncflow`
- `go test ./...`

结果均通过。

## 6. 遗留事项

- 未做桌面端手工点击验证；本次以本地配置归一化和云端 manifest 选择回归测试作为主要证明。
