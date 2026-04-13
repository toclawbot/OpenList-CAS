<div align="center">
  <img style="width: 128px; height: 128px;" src="https://raw.githubusercontent.com/OpenListTeam/Logo/main/logo.svg" alt="logo" />

  <p><em>OpenList 是一个有韧性、长期治理、社区驱动的 AList 分支，旨在防御基于信任的开源攻击。</em></p>

  <img src="https://goreportcard.com/badge/github.com/OpenListTeam/OpenList/v3" alt="latest version" />
  <a href="https://github.com/OpenListTeam/OpenList/blob/main/LICENSE"><img src="https://img.shields.io/github/license/OpenListTeam/OpenList" alt="License" /></a>
  <a href="https://github.com/OpenListTeam/OpenList/actions?query=workflow%3ABuild"><img src="https://img.shields.io/github/actions/workflow/status/OpenListTeam/OpenList/build.yml?branch=main" alt="Build status" /></a>
  <a href="https://github.com/OpenListTeam/OpenList/releases"><img src="https://img.shields.io/github/release/OpenListTeam/OpenList" alt="latest version" /></a>

  <a href="https://github.com/OpenListTeam/OpenList/discussions"><img src="https://img.shields.io/github/discussions/OpenListTeam/OpenList?color=%23ED8936" alt="discussions" /></a>
  <a href="https://github.com/OpenListTeam/OpenList/releases"><img src="https://img.shields.io/github/downloads/OpenListTeam/OpenList/total?color=%239F7AEA&logo=github" alt="Downloads" /></a>
</div>

---
# OpenList-CAS

基于 [OpenList](https://github.com/OpenListTeam/OpenList) 的增强分支，围绕 `.cas` 秒传元数据工作流进行优化，实现**低存储占用 + 快速恢复文件**的高效方案。

---

## ✨ TL;DR

* 📦 上传文件 → 自动生成 `.cas` 元数据
* 🗑️ 可删除原文件，仅保留 `.cas` 节省空间
* ⚡ 通过 `.cas` 可秒传恢复原文件（无需重新上传）

---

## 📑 目录

- [🚀 使用场景](#-使用场景)
- [🔄 工作流程](#-工作流程)
- [🔧 核心特性](#-核心特性)
- [📦 支持驱动](#-支持驱动)
- [⚙️ 配置说明](#️-配置说明)
- [🏷️ 命名规则](#️-命名规则)
- [🖥️ 本地存储说明](#️-本地存储说明local)
- [🐳 部署指南](#-部署指南)
- [🌐 访问](#-访问)
- [⚠️ 常见问题](#️-常见问题)
- [🔗 与上游项目](#-与上游项目)
- [📜 免责声明](#-免责声明)
- [🙏 致谢](#-致谢)

---

## 🚀 使用场景

* 📉 **低存储环境（VPS / NAS）**
  只保存 `.cas`，极大减少空间占用

* ☁️ **网盘秒传优化**
  利用哈希直接恢复文件，避免重复上传

* 🎬 **媒体库归档**
  平时只存元数据，需要时再恢复原文件

* 🔁 **自动化工作流**
  监控 `.cas` 文件并自动恢复内容

---

## 🔄 工作流程

```text
上传文件 → 生成 .cas → （可选）删除原文件 / 上传 .cas → 秒传恢复原文件
```

---

## 🔧 核心特性

* 支持将普通文件转换为 `.cas` 元数据文件
* 支持“生成后删除源文件”的轻量存储模式
* 支持通过 `.cas` 秒传恢复文件（非上传 `.cas` 本身）
* 支持重命名 `.cas` 后恢复（自动补全扩展名）
* 支持自动监控目录并恢复 `.cas` 文件

---

## 📦 支持驱动

| 驱动           | 支持情况          |
| ------------ | ------------- |
| `189Cloud`   | ✅ 完整支持        |
| `189CloudPC` | ✅ 完整支持        |
| `Local`      | ⚠️ 仅支持生成 / 删除 |

---

## ⚙️ 配置说明

| 配置项                             | 默认值 | 适用驱动     | 说明              |
| ------------------------------- | --- | -------- | --------------- |
| Generate cas                    | ❌   | 全部       | 上传后生成 `.cas`    |
| Delete source                   | ❌   | 全部       | 生成后删除原文件        |
| Restore source from cas         | ❌   | 189Cloud | 上传 `.cas` 时恢复文件 |
| Restore source use current name | ❌   | 189Cloud | 使用当前文件名恢复       |
| Delete CAS after restore        | ❌   | 189Cloud | 恢复后删除 `.cas`    |
| Auto restore existing cas       | ❌   | 189Cloud | 自动监听恢复          |
| Auto restore existing cas paths | -   | 189Cloud | 监听目录            |

---


### 👉 低存储模式（推荐）

开启：

* ✅ Generate cas
* ✅ Delete source

效果：

```text
movie.mp4 → movie.mp4.cas
（（原文件删除，保留 .cas））
```

---

## 🏷️ 命名规则

开启 **“使用当前文件名恢复”** 时：

| 操作          | 恢复结果              |
| ----------- | ----------------- |
| `a.mp4.cas` | → `a.mp4`         |
| `a.cas`     | → `a.mp4`（自动补扩展名） |

关闭该选项时：

* 优先使用 `.cas` 内记录的原始文件名

---

## 🖥️ 本地存储说明（Local）

支持：

* 生成 `.cas`
* 删除源文件

暂不支持：

* 秒传恢复

---

## 🐳 部署指南

### Docker

```bash
docker run -d --restart=unless-stopped \
  -v /etc/openlist:/opt/openlist/data \
  -p 5244:5244 \
  -e PUID=0 \
  -e PGID=0 \
  -e UMASK=022 \
  --name="openlist-cas" \
  freeyua/openlist-cas:latest
```

---

### Docker Compose

```yaml
services:
  openlist-cas:
    image: freeyua/openlist-cas:latest
    container_name: openlist-cas
    restart: unless-stopped
    ports:
      - "5244:5244"
    volumes:
      - ./data:/opt/openlist/data
    environment:
      - PUID=0
      - PGID=0
      - UMASK=022
```

---

## 🌐 访问

启动后访问：

```
http://localhost:5244
```

---

## ⚠️ 常见问题

### ❗ 无法恢复文件

* 驱动不支持秒传能力

### ❗ 上传 `.cas` 没反应

* 未开启 `Restore source from cas`

### ❗ 文件名不正确

* 检查 `Restore source use current name`

---

## 🔗 与上游项目

* 上游项目：OpenList
* 基线版本：v4.2.1
* 本项目为非官方增强分支

---

## 📜 免责声明

1. 本项目仅用于学习与技术研究，请遵守相关法律法规，请勿用于商业用途。
2. 本项目所涉及的任何脚本、程序或资源，仅用于测试和研究目的。
3. 使用者应在下载后的24小时内删除相关文件。
4. 使用者需自行承担使用本项目可能产生的一切法律后果和风险，作者不承担任何责任。
5. 如果您不能接受本声明的任何条款，请立即停止使用本项目。

---

## 🙏 致谢

感谢原项目 [OpenList](https://github.com/OpenListTeam/OpenList) 提供的基础能力。
