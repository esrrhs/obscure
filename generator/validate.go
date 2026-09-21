package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// sceneNum 用于扫描 content/scenes 目录里的数字子目录名。image.go 依赖它。
var sceneNum = regexp.MustCompile(`^\d+$`)

// validateSceneText 校验一幕的 text.txt 是否合法。
// 用于 LLM 出稿后判定是否要重试。
// 合法条件：非空、无英文字母、至少 80 汉字、无 LLM 元话语。
func validateSceneText(dir string) (bool, string) {
	p := filepath.Join(dir, "text.txt")
	if !nonEmpty(p) {
		return false, "text.txt 缺失或空"
	}
	b, _ := os.ReadFile(p)
	s := strings.TrimSpace(string(b))
	if hasEnglishLetters(s) {
		return false, "text.txt 包含英文字母"
	}
	if runeCount := utf8.RuneCountInString(s); runeCount < 80 {
		return false, fmt.Sprintf("text.txt 过短 (%d 字，可能是 LLM 短路)", runeCount)
	}
	if bad := detectMetaTalk(s); bad != "" {
		return false, fmt.Sprintf("text.txt 含 LLM 元话语: %q", bad)
	}
	return true, ""
}

// metaTalkPatterns 是 LLM 短路 / 拒答 / 自我暴露的常见短语。
var metaTalkPatterns = []string{
	"作为AI", "作为一个AI", "作为语言模型", "作为人工智能",
	"我是AI", "我是一个AI",
	"抱歉，", "很抱歉", "对不起，",
	"无法完成", "无法继续", "无法生成", "无法提供",
	"已取消", "已跳过", "有其他需要", "需要帮忙",
	"这是虚构", "这是一个虚构", "让我为您", "让我为你",
	"接下来我将",
}

func detectMetaTalk(s string) string {
	lower := strings.ToLower(s)
	for _, kw := range metaTalkPatterns {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return kw
		}
	}
	return ""
}

func fixEscapedQuotes(s string) string    { return strings.ReplaceAll(s, `\"`, `"`) }
func stripMarkdownAsterisks(s string) string { return strings.ReplaceAll(s, "**", "") }

func hasEnglishLetters(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func nonEmpty(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return fi.Size() > 0
}

// cleanLLMText 对 LLM 返回的正文做一次统一清理：去转义引号、去 Markdown 加粗。
func cleanLLMText(s string) string {
	return stripMarkdownAsterisks(fixEscapedQuotes(strings.TrimSpace(s)))
}

// 供 image.go 里 strconv 用
var _ = strconv.Itoa
