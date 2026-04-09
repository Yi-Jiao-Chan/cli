---
name: lark-slides
version: 1.0.0
description: "飞书幻灯片：以XML格式创建和管理PPT。创建空白演示文稿、读取PPT全文信息、创建和删除幻灯片页面。当用户需要创建PPT、读取PPT内容、管理幻灯片页面时使用。"
metadata:
  requires:
    bins: ["lark-cli"]
  cliHelp: "lark-cli slides --help"
---

# slides (v1)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../lark-shared/SKILL.md`](../lark-shared/SKILL.md)，其中包含认证、权限处理**

## 执行前必做

> **重要**：`references/slides_xml_schema_definition.xml` 是此 skill 唯一正确的 XML 协议来源；其他 md 仅是对它和 CLI schema 的摘要。

| 命令 | 必须先阅读的文档 |
|------|-----------------|
| `xml_presentations.create` | [lark-slides-xml-presentations-create.md](references/lark-slides-xml-presentations-create.md) |
| `xml_presentations.get` | [lark-slides-xml-presentations-get.md](references/lark-slides-xml-presentations-get.md) |
| `xml_presentation.silde.create` | [lark-slides-xml-presentation-slides-create.md](references/lark-slides-xml-presentation-slides-create.md) |
| `xml_presentation.silde.delete` | [lark-slides-xml-presentation-slides-delete.md](references/lark-slides-xml-presentation-slides-delete.md) |

**涉及 XML 格式时，阅读顺序：**
1. [xml-schema-quick-ref.md](references/xml-schema-quick-ref.md) — **首选：Schema 精简速查**
2. [xml-format-guide.md](references/xml-format-guide.md) — 详细结构、属性与示例
3. [examples.md](references/examples.md) — CLI 调用示例
4. [slides_demo.xml](references/slides_demo.xml) — 真实 XML 示例
5. [slides_xml_schema_definition.xml](references/slides_xml_schema_definition.xml) — 完整 Schema

## 核心概念

### URL 格式与 Token

| URL 格式 | 示例 | Token 类型 | 处理方式 |
|----------|------|-----------|----------|
| `/slides/` | `https://example.larkoffice.com/slides/xxxxxxxxxxxxx` | `xml_presentation_id` | URL 路径中的 token 直接作为 `xml_presentation_id` 使用 |
| `/wiki/` | `https://example.larkoffice.com/wiki/wikcnxxxxxxxxx` | `wiki_token` | ⚠️ **不能直接使用**，需要先查询获取真实的 `obj_token` |

### Wiki 链接特殊处理（关键！）

知识库链接（`/wiki/TOKEN`）背后可能是云文档、电子表格、幻灯片等不同类型的文档。**不能直接假设 URL 中的 token 就是 `xml_presentation_id`**，必须先查询实际类型和真实 token。

#### 处理流程

1. **使用 `wiki.spaces.get_node` 查询节点信息**
   ```bash
   lark-cli wiki spaces get_node --params '{"token":"wiki_token"}'
   ```

2. **从返回结果中提取关键信息**
   - `node.obj_type`：文档类型，幻灯片对应 `slides`
   - `node.obj_token`：**真实的演示文稿 token**（用于后续操作）
   - `node.title`：文档标题

3. **确认 `obj_type` 为 `slides` 后，使用 `obj_token` 作为 `xml_presentation_id`**

#### 查询示例

```bash
# 查询 wiki 节点
lark-cli wiki spaces get_node --params '{"token":"OFG3w29CWiB0xNkVvhEcC2ynnAg"}'
```

返回结果示例：
```json
{
   "node": {
      "obj_type": "slides",
      "obj_token": "CaABs8G8Kl5UoDd9y7xcwjz9ndd",
      "title": "2026 产品年度总结",
      "node_type": "origin",
      "space_id": "7028488849126932483"
   }
}
```

```bash
# 用 obj_token 读取幻灯片内容
lark-cli slides xml_presentations get --params '{"xml_presentation_id":"CaABs8G8Kl5UoDd9y7xcwjz9ndd"}'
```

### 资源关系

```
Wiki Space (知识空间)
└── Wiki Node (知识库节点, obj_type: slides)
    └── obj_token → xml_presentation_id

Slides (演示文稿)
├── xml_presentation_id (演示文稿唯一标识)
├── revision_id (版本号)
└── Slide (幻灯片页面)
    └── slide_id (页面唯一标识)
```

## API Resources

```bash
lark-cli schema slides.<resource>.<method>    # 调用 API 前必须先查看参数结构
lark-cli slides <resource> <method> [flags]  # 调用 API
```

> **重要**：使用原生 API 时，必须先运行 `schema` 查看 `--data` / `--params` 参数结构，不要猜测字段格式。

### xml_presentations

  - `create` — 创建空白 PPT（当前仅支持标题和长宽）
  - `get` — 读取ppt全文信息，xml格式返回

### xml_presentation.silde

  - `create` — 在指定 xml 演示文稿下创建页面
  - `delete` — 删除指定 xml 演示文稿下的页面

## 意图 → 命令索引

| 意图 | 推荐命令 | 备注 |
|------|---------|------|
| 创建空白 PPT | `lark-cli slides xml_presentations create` | 当前仅用于创建空白 PPT，建议传 `<presentation ...><title>...</title></presentation>` |
| 读取 PPT 内容 | `lark-cli slides xml_presentations get` | `--params` 传入 `{"xml_presentation_id":"..."}` |
| 添加幻灯片页面 | `lark-cli slides xml_presentation.silde create` | 创建空白 PPT 后，用它逐页添加 slide |
| 删除幻灯片页面 | `lark-cli slides xml_presentation.silde delete` | `--params` 传入 `xml_presentation_id` 和 `slide_id` |

## 核心规则

1. **先查 schema**：调用前先运行 `lark-cli schema slides.<resource>.<method>`
2. **命名空间建议**：协议标准写法应带 `xmlns`，例如 `<presentation xmlns="http://www.larkoffice.com/sml/2.0" ...>`；当前服务端实现可能兼容不带 `xmlns` 的输入，但不作为协议保证
3. **根结构固定**：`<presentation>` 直接子元素只有 `<title>`、`<theme>`、`<slide>`
4. **slide 结构固定**：`<slide>` 直接子元素只有 `<style>`、`<data>`、`<note>`
5. **文本通过 content 表达**：页面正文通常放在 `shape/table/note` 内的 `<content>` 中
6. **创建流程要分两步**：先用 `xml_presentations.create` 创建空白 PPT，再用 `xml_presentation.silde.create` 逐页添加 slide
7. **保存关键 ID**：后续操作需要 `xml_presentation_id`、`slide_id`、`revision_id`
8. **删除谨慎**：删除操作不可逆，且至少保留一页幻灯片

## 权限表

| 方法 | 所需 scope |
|------|-----------|
| `xml_presentations.create` | `slides:presentation:create` |
| `xml_presentations.get` | `slides:presentation:read` |
| `xml_presentation.silde.create` | `slides:presentation:update` 或 `slides:presentation:write_only` |
| `xml_presentation.silde.delete` | `slides:presentation:update` 或 `slides:presentation:write_only` |

## 常见错误速查

| 错误码 | 含义 | 解决方案 |
|--------|------|----------|
| 400 | XML 格式错误 | 检查 XML 语法，确保标签闭合 |
| 400 | create 内容超出支持范围 | `xml_presentations.create` 仅用于创建空白 PPT，不要在这里传完整 slide 内容 |
| 400 | 请求包装错误 | 检查 `--data` 是否按 schema 传入 `xml_presentation.content` 或 `slide.content` |
| 404 | 演示文稿不存在 | 检查 `xml_presentation_id` 是否正确 |
| 404 | 幻灯片不存在 | 检查 `slide_id` 是否正确 |
| 403 | 权限不足 | 检查是否拥有对应的 scope |
| 400 | 无法删除唯一幻灯片 | 演示文稿至少保留一页幻灯片 |

## 参考文档

### 快速参考
- [xml-schema-quick-ref.md](references/xml-schema-quick-ref.md) — **XML Schema 精简速查**
- [xml-format-guide.md](references/xml-format-guide.md) — XML 结构、内容模型、常用元素
- [examples.md](references/examples.md) — 常见 CLI 调用示例
- [slides_demo.xml](references/slides_demo.xml) — 真实 XML 示例

### 命令参考

- [lark-slides-xml-presentations-create.md](references/lark-slides-xml-presentations-create.md) — 创建空白 PPT
- [lark-slides-xml-presentations-get.md](references/lark-slides-xml-presentations-get.md) — 读取 PPT
- [lark-slides-xml-presentation-slides-create.md](references/lark-slides-xml-presentation-slides-create.md) — 添加幻灯片
- [lark-slides-xml-presentation-slides-delete.md](references/lark-slides-xml-presentation-slides-delete.md) — 删除幻灯片

### Schema 定义

- [xml-schema-quick-ref.md](references/xml-schema-quick-ref.md) — Schema 精简参考
- [slides_xml_schema_definition.xml](references/slides_xml_schema_definition.xml) — **完整 Schema 定义**（唯一协议依据）

> **注意**：如果 md 内容与 `slides_xml_schema_definition.xml` 或 `lark-cli schema slides.<resource>.<method>` 输出不一致，以后两者为准。
