package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// LLM 黑盒接口。
type LLM interface {
	Complete(prompt string) (string, error)
}

// -----------------------------------------------------------------------------
// MimoLLM: 通过 mimo run 调用 AI
// -----------------------------------------------------------------------------

type MimoLLM struct{}

func NewMimoLLM() *MimoLLM {
	return &MimoLLM{}
}

// 过滤 mimo 输出中的状态行，只保留真正的内容。
// mimo 会输出如下噪声行：
//   - 空行（\r\n）
//   - "> build · mimo-auto"
//   - "⚙ skill_search ..."
//   - 其他以 "> " 或 "⚙ " 开头的 UI 状态行

func filterMimoOutput(raw string) string {
	// 统一换行符
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")

	var kept []string
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		// 跳过空行（noise 行之间的空行）和状态行
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "> ") {
			continue
		}
		if strings.HasPrefix(trimmed, "⚙ ") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func (m *MimoLLM) Complete(prompt string) (string, error) {
	// 无限重试，指数退退（上限 60s）
	const maxWait = 60 * time.Second
	wait := 2 * time.Second

	for attempt := 1; ; attempt++ {
		task := fmt.Sprintf(
			"请直接输出以下任务的结果文本，不要使用任何工具，不要写文件，不要做额外操作，只输出纯文本回答：\n\n%s",
			prompt,
		)

		cmd := exec.Command("mimo", "run", "--dangerously-skip-permissions", task)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		var lastErr error
		if err := cmd.Run(); err != nil {
			lastErr = fmt.Errorf("mimo 执行失败: %w, stderr: %s", err, stderr.String())
		} else {
			result := filterMimoOutput(stdout.String())
			if result == "" {
				lastErr = fmt.Errorf("mimo 输出为空, stdout=%q, stderr=%q", stdout.String(), stderr.String())
			} else {
				return result, nil
			}
		}

		fmt.Fprintf(os.Stderr, "  [第 %d 次失败，%v 后重试] %v\n", attempt, wait, lastErr)
		time.Sleep(wait)
		wait *= 2
		if wait > maxWait {
			wait = maxWait
		}
	}
}

// -----------------------------------------------------------------------------
// LongCatLLM: 通过美团 LongCat OpenAI 兼容接口调用
// -----------------------------------------------------------------------------

type LongCatLLM struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

func NewLongCatLLM(apiKey string) *LongCatLLM {
	if apiKey == "" {
		apiKey = os.Getenv("LONGCAT_API_KEY")
	}
	return &LongCatLLM{
		endpoint: "https://api.longcat.chat/openai/v1/chat/completions",
		apiKey:   apiKey,
		model:    "LongCat-2.0",
		client:   &http.Client{Timeout: 5 * time.Minute},
	}
}

type longCatReq struct {
	Model    string           `json:"model"`
	Messages []longCatMessage `json:"messages"`
}

type longCatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type longCatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (l *LongCatLLM) Complete(prompt string) (string, error) {
	if l.apiKey == "" {
		return "", fmt.Errorf("未设置 LONGCAT_API_KEY 环境变量（或未通过 -longcat-key 传入）")
	}
	const maxWait = 60 * time.Second
	wait := 2 * time.Second

	body, err := json.Marshal(longCatReq{
		Model:    l.model,
		Messages: []longCatMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", err
	}

	for attempt := 1; ; attempt++ {
		var lastErr error

		req, err := http.NewRequest("POST", l.endpoint, bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+l.apiKey)

		resp, err := l.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("请求失败: %w", err)
		} else {
			raw, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != 200 {
				lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
			} else {
				var out longCatResp
				if err := json.Unmarshal(raw, &out); err != nil {
					lastErr = fmt.Errorf("解析响应失败: %w, body=%s", err, string(raw))
				} else if out.Error != nil {
					lastErr = fmt.Errorf("API 错误: %s", out.Error.Message)
				} else if len(out.Choices) == 0 || strings.TrimSpace(out.Choices[0].Message.Content) == "" {
					lastErr = fmt.Errorf("响应内容为空, body=%s", string(raw))
				} else {
					return strings.TrimSpace(out.Choices[0].Message.Content), nil
				}
			}
		}

		fmt.Fprintf(os.Stderr, "  [第 %d 次失败，%v 后重试] %v\n", attempt, wait, lastErr)
		time.Sleep(wait)
		wait *= 2
		if wait > maxWait {
			wait = maxWait
		}
	}
}

// -----------------------------------------------------------------------------
// MockLLM: 用于流程调试的假 LLM
// -----------------------------------------------------------------------------

type MockLLM struct {
	counter int
}

func NewMockLLM() *MockLLM { return &MockLLM{} }

func (m *MockLLM) Complete(prompt string) (string, error) {
	m.counter++
	switch {
	case strings.Contains(prompt, "写游戏的第 1 幕"):
		return "崇祯十六年冬夜，乾清宫烛火摇曳。你猛地睁开眼，触目所及是明黄色的帷幔与漆金龙柱。" +
			"手指下是冰凉的奏折，鼻尖萦绕着龙涎香和一丝挥之不去的血腥气。殿外传来更漏声，" +
			"一名太监匍匐在地颤声禀报：\"陛下，李自成部已破潼关，孙传庭战殁……\"" +
			"你的心跳如擂鼓——这不是梦。史书上那个吊死煤山的皇帝，此刻就是你自己。", nil

	case strings.Contains(prompt, "从本幕结尾抽出"):
		if m.counter > 30 {
			return "ENDING: 时局至此已定，你在这一刻做出了最后的抉择，历史在此刻分岔。", nil
		}
		return fmt.Sprintf(
			"EVENT: 殿内群臣屏息等待你的旨意，窗外风雪骤起，此刻的每一个决定都关乎社稷存亡。\n"+
				"CHOICE1: 立即召孙承宗等老臣入宫议事（mock-%d-a）\n"+
				"CHOICE2: 秘密派人接触李自成谈判（mock-%d-b）\n"+
				"CHOICE3: 下诏南迁，避其锋芒（mock-%d-c）",
			m.counter, m.counter, m.counter,
		), nil

	case strings.Contains(prompt, "写下一幕的剧情正文"):
		return "你颁下旨意，殿内一时鸦雀无声，随即群臣纷纷领命而去。你独坐龙椅之上，" +
			"望着窗外风雪，深知这一步一旦踏出便再无回头。片刻之后，新的奏报又送到了案前……", nil

	case strings.Contains(prompt, "产出更新后的故事记忆"):
		return "主角穿越到崇祯身上，面临李自成东进、清军窥伺的危局，在关键节点做出了抉择。", nil
	}

	return "[MOCK: 未识别的 prompt 类型]", nil
}
