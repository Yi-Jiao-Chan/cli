---
name: lark-whiteboard
description: >
  飞书画板：查询和编辑飞书云文档中的画板。支持导出画板为预览图片、导出原始节点结构、使用 PlantUML/Mermaid 代码��生格式��生格式更新画板内容。
  附有详细的画板绘制工作流程指南，从路由决策到交付审查。
compatibility: Requires Node.js 18+
metadata:
  requires:
    bins: [ "lark-cli" ]
---

# whiteboard (v1)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../lark-shared/SKILL.md`](../lark-shared/SKILL.md)，其中包含认证、权限处理**
**CRITICAL — 绘制画板务必按照 [画板绘制工作流程](workflow.md) 中的流程步骤操作**

> [!NOTE]
> **环境依赖**：绘制画板需要 `@larksuite/whiteboard-cli`（画板 Node.js CLI 工具），以及 `lark-cli`（LarkSuite CLI 工具）。
> 如果执行失败，手动安装后重试：`npm install -g @larksuite/whiteboard-cli@^0.2.0`

> [!IMPORTANT]
> 执行 `npm install` 安装新的依赖前，务必征得用户同意！
> ## 快速决策

当需要在飞书文档中插入图表时：

1. 能否使用飞书画板？

- 能 → 走画板路径，**MUST** 遵循 [画板绘制工作流程](workflow.md) 中的流程步骤（推荐！可编辑、可协作）
- 不能 → 走图片路径，不使用本技能

| 用户需求                             | 推荐 Shortcut                                                                                                         |
|----------------------------------|---------------------------------------------------------------------------------------------------------------------|
| "查看这个画板的内容"                      | [`+query --output_as image`](references/lark-whiteboard-query.md)                                                   |
| "导出画板为图片"                        | [`+query --output_as image`](references/lark-whiteboard-query.md)                                                   |
| "获取画板的 PlantUML/Mermaid 代码"      | [`+query --output_as code`](references/lark-whiteboard-query.md)                                                    |
| "检查画板是否由 PlantUML/Mermaid 代码块组成" | [`+query --output_as code`](references/lark-whiteboard-query.md)                                                    |
| "修改画板某个节点的颜色或文字"                 | [`+query --output_as raw`](references/lark-whiteboard-query.md) 后 [`+update`](references/lark-whiteboard-update.md) |
| "用 PlantUML 绘制画板"                | [`+update --input_format plantuml`](references/lark-whiteboard-update.md)                                           |
| "用 Mermaid 绘制画板"                 | [`+update --input_format mermaid`](references/lark-whiteboard-update.md)                                            |
| "在画板绘制复杂图表"                      | 参考 [workflow.md](workflow.md)                                                                                       |

## Shortcuts

| Shortcut                                          | 说明                                          |
|---------------------------------------------------|---------------------------------------------|
| [`+query`](references/lark-whiteboard-query.md)   | 查询画板，导出为预览图片、代码或原始节点结构                      |
| [`+update`](references/lark-whiteboard-update.md) | 更新画板内容，支持 PlantUML、Mermaid 或 OpenAPI 原生格式输入 |

## 目录

### 核心流程

- [workflow.md](workflow.md) - 画板创作完整工作流程，从路由决策到交付审查

### Lark CLI 指令

- [references/lark-whiteboard-query.md](references/lark-whiteboard-query.md) - `+query` 查询画板：导出图片、提取代码、获取原始节点
- [references/lark-whiteboard-update.md](references/lark-whiteboard-update.md) - `+update` 更新画板：支持
  PlantUML/Mermaid/原生格式

### 核心参考模块（绘制画板必读）

- [references/schema.md](references/schema.md) - DSL 语法规范：节点类型、属性、尺寸值
- [references/content.md](references/content.md) - 内容规划：信息量匹配、分组策略、连线预判
- [references/layout.md](references/layout.md) - 布局系统：Flex/Dagre/绝对定位决策、网格方法论
- [references/style.md](references/style.md) - 配色系统：色板选择、分层上色、视觉层级
- [references/typography.md](references/typography.md) - 排版规则：字号层级、对齐方式、图文组合
- [references/connectors.md](references/connectors.md) - 连线系统：拓扑规划、自动绕线、锚点选择

### 场景指南（按图表类型选读）

- [scenes/architecture.md](scenes/architecture.md) - 架构图：分层架构、微服务架构
- [scenes/flowchart.md](scenes/flowchart.md) - 流程图：业务流、状态机、条件判断链路
- [scenes/organization.md](scenes/organization.md) - 组织架构图：公司组织、树形层级
- [scenes/milestone.md](scenes/milestone.md) - 里程碑/时间线
- [scenes/fishbone.md](scenes/fishbone.md) - 鱼骨图：因果分析、根因分析
- [scenes/comparison.md](scenes/comparison.md) - 对比图：方案对比、功能矩阵
- [scenes/flywheel.md](scenes/flywheel.md) - 飞轮图：增长飞轮、闭环链路
- [scenes/pyramid.md](scenes/pyramid.md) - 金字塔图：层级结构、需求层次
- [scenes/bar-chart.md](scenes/bar-chart.md) - 柱状图
- [scenes/line-chart.md](scenes/line-chart.md) - 折线图
- [scenes/treemap.md](scenes/treemap.md) - 树状图：矩形树图、层级占比
- [scenes/funnel.md](scenes/funnel.md) - 漏斗图：转化漏斗、销售漏斗
- [scenes/swimlane.md](scenes/swimlane.md) - 泳道图：跨角色流程、端到端链路
- [scenes/mermaid.md](scenes/mermaid.md) - 使用 Mermaid 创建以下类型图表：思维导图、时序图、类图、饼图
