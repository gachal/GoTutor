# 添加章节

> [English](ADDING_A_CHAPTER.md) | 中文

一个章节是一个独立的学习单元，包含：
- **模板**（template）—— 给学习者看的 Go 骨架，带 `// TODO` 标记
- **提示**（hints）—— 每个 TODO 行的中英（zh + en）双语文本
- **测试**（tests）—— 学习者代码必须通过的 `go test` 用例
- **答案**（solution）—— 参考答案，按需通过 `GET /api/chapters/:id/solution`
  端点暴露给学习者（在章节详情页的"推荐答案"弹窗里展示）

本文档以添加第 3 章（一个 `time` 包练习）为例，走完整个流程。

## 1. 选定 ID 和概念

两字母的 ID 比较好。选一个本章要教的 Go 概念：
- `time` —— `time.Now`、`time.Sleep`、`time.Duration`
- `sort` —— `sort.Slice`、自定义 `Less`
- `io` —— `io.Reader` / `io.Writer` 组合

## 2. 创建内容目录

```
backend/chapters/content/<id>/
├── template.txt              # 带 // TODO 的骨架
├── hints.yaml                # 按行号索引的双语提示
├── solution/main.txt         # 参考答案
└── tests/<name>_test.go.txt  # 测试文件（后缀必须是 .go.txt）
```

`go:embed` 会跳过 `.go` 文件——这就是模板和测试以 `.txt` 形式分发的
原因。测试的 `.go.txt` 后缀在提交时还原为 `_test.go`，让 `go test` 能
发现它们。

## 3. 写模板

以一行 `// Chapter N: <title>` 头开头，然后是 `package main` 和函数桩。
用 TODO 标出空缺：

```go
// Chapter 3: Time Wrapper — initial skeleton.
package main

import (
	"fmt"
	"time"
)

// FormatDuration returns a human string for a Duration, e.g.
// "1h2m3s" → "1 hour 2 minutes 3 seconds".
func FormatDuration(d time.Duration) string {
	// TODO 1: convert d to hours, minutes, seconds.
	return ""
}

func main() {
	fmt.Println(FormatDuration(1 * time.Hour))
}
```

## 4. 写提示

`hints.yaml` 按对应 `// TODO` 的 1-based 行号索引提示文本：

```yaml
todos:
  - line: 12
    hint:
      zh: 用 d.Hours(), d.Minutes() 等方法提取各部分。
      en: Use d.Hours(), d.Minutes() etc. to extract each part.
```

用 `grep -n '// TODO' template.txt` 确认行号；后端的 TODO 扫描用的是
`^\s*//\s*TODO\b`，所以任何注释标记字面上以 `// TODO` 开头的行都会被
标记。

## 5. 写测试

测试必须确定且快速（<10 秒——校验器的超时）。需要服务器或外部状态的
章节用 `httptest`（见 `urlcheck/tests/urlcheck_test.go.txt`）。纯函数
最适合表驱动测试。

```go
// tests/time_test.go.txt
package main

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{1 * time.Hour, "1 hour"},
		{1*time.Hour + 30*time.Minute, "1 hour 30 minutes"},
	}
	for _, tc := range cases {
		got := FormatDuration(tc.in)
		if got != tc.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
```

## 6. 写答案

把参考答案放进 `solution/main.txt`。它会按需经
`GET /api/chapters/:id/solution`（章节详情页的"推荐答案"弹窗）提供给
学习者；CI 也可以编译它 + 跑测试，验证本章可解。

## 7. 注册章节

编辑 `backend/chapters/registry.go`，往 `registry` 切片追加：

```go
{
	ID:          "time",
	Title:       Locale{Zh: "时间格式化", En: "Time Formatting"},
	Description: Locale{
		Zh: "用 time 包将 Duration 转成人类可读的字符串。",
		En: "Convert a time.Duration into a human-readable string.",
	},
	Ordinal:      3,
	// 可选：在安全标准库基线之外扩展校验器白名单。
	// time/fmt 等已默认允许；只有 net/http 之类才需要。
	// AllowImports: []string{"net/http"},
	contentDir:   "time",
},
```

所有章节从一开始就全部解锁——学习者可自由探索任意章节（见
`internal/api/chapters.go` 里的 `HandleListChapters`，其中 `Unlocked`
恒为 `true`）。`Ordinal` 仅控制显示顺序和排序。`Completed` 仍反映用户
是否曾通过。

## 8. 端到端测试

```bash
# 重建后端（会内嵌新的 content/）
make backend-dev

# 另一个终端——验证章节出现在列表里：
curl localhost:8081/api/chapters

# 获取模板：
curl localhost:8081/api/chapters/time/template

# 测试提示查询：
curl 'localhost:8081/api/chapters/time/hint?line=12'

# 提交参考答案：
SOLUTION=$(cat backend/chapters/content/time/solution/main.txt | \
           python3 -c "import sys,json; print(json.dumps({'userCode': sys.stdin.read()}))")
curl -X POST localhost:8081/api/chapters/time/submit \
  -H 'Content-Type: application/json' -d "$SOLUTION"
```

返回应当是 `passed: true`，章节的 `completed` 标志应翻转（所有章节
本就从头解锁）。

## 9. 验证 AST 策略

校验器的默认白名单在 `backend/internal/verifier/policy.go`
（`SafeStdlibImports` + `DeniedStdlibImports`）。若你的章节需要一个
不在任一列表里的导入，把它加到该章的 `AllowImports`。测试拒绝：

```bash
# 提交带禁用导入的代码——应在 AST 阶段被拒。
BAD=$(python3 -c "import json; print(json.dumps({'userCode': 'package main\nimport \"os/exec\"\nfunc main() {}\n'}))")
curl -X POST localhost:8081/api/chapters/time/submit \
  -H 'Content-Type: application/json' -d "$BAD"
# 期望：{passed: false, output: 'import "os/exec" is not allowed...'}
```

## 10. 文档

把章节加到 README 的章节列表里。完成。
