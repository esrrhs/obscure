package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// -----------------------------------------------------------------------------
// 引擎接口：所有生图引擎实现 name/ext/dstPath/genOnce 四个方法即可。
// -----------------------------------------------------------------------------

type imageEngine interface {
	name() string      // 用于日志显示
	ext() string       // 输出文件扩展名（不含点）
	needCensorSkip(error) bool // 该错误是不是内容审核类（永久失败，跳过）
	genOnce(prompt, dstPath string) error
}

// -----------------------------------------------------------------------------
// 引擎一：CogView-3-Flash（智谱免费文生图，OpenAI 风格）
//   - 内容审核严格（code=1301 永久拦截）
//   - 分辨率 1344x768
//   - 输出 JPEG
// -----------------------------------------------------------------------------

var cogViewAPIKey = os.Getenv("COGVIEW_API_KEY")

const cogViewEndpoint = "https://open.bigmodel.cn/api/paas/v4/images/generations"
const cogViewModel = "cogview-3-flash"
const cogViewSize = "1344x768"

type cogViewReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Size   string `json:"size"`
}

type cogViewResp struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ContentFilter []struct {
		Level int    `json:"level"`
		Role  string `json:"role"`
	} `json:"contentFilter"`
}

type cogViewEngine struct{}

func (cogViewEngine) name() string { return "CogView" }
func (cogViewEngine) ext() string  { return "jpg" }
func (cogViewEngine) needCensorSkip(err error) bool {
	return err != nil && strings.Contains(err.Error(), "1301")
}

func (cogViewEngine) genOnce(prompt, dstPath string) error {
	if cogViewAPIKey == "" {
		return fmt.Errorf("未设置 COGVIEW_API_KEY 环境变量")
	}
	body, err := json.Marshal(cogViewReq{Model: cogViewModel, Prompt: prompt, Size: cogViewSize})
	if err != nil {
		return fmt.Errorf("编码请求失败: %w", err)
	}
	req, err := http.NewRequest("POST", cogViewEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cogViewAPIKey)

	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var out cogViewResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return fmt.Errorf("解析响应失败: %w, body=%s", err, string(raw))
	}
	if out.Error != nil {
		return fmt.Errorf("API 错误 %s: %s", out.Error.Code, out.Error.Message)
	}
	if len(out.Data) == 0 || out.Data[0].URL == "" {
		return fmt.Errorf("响应无图片 URL, body=%s", string(raw))
	}
	return downloadTo(out.Data[0].URL, dstPath, client)
}

// -----------------------------------------------------------------------------
// 引擎二：Kwai-Kolors on SiliconFlow（免费，中文提示词优秀，无严格内容审核）
//   - 分辨率 2016x1120（16:9，Kolors 支持的极限尺寸，<2048）
//   - 输出 PNG
// -----------------------------------------------------------------------------

var kolorsAPIKey = os.Getenv("KOLORS_API_KEY")

const kolorsEndpoint = "https://api.siliconflow.cn/v1/images/generations"
const kolorsModel = "Kwai-Kolors/Kolors"
const kolorsSize = "2016x1120"
const kolorsSteps = 30
const kolorsGuidance = 7.5

type kolorsReq struct {
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt"`
	ImageSize      string  `json:"image_size"`
	NumSteps       int     `json:"num_inference_steps"`
	GuidanceScale  float64 `json:"guidance_scale"`
}

type kolorsResp struct {
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type kolorsEngine struct{}

func (kolorsEngine) name() string          { return "Kolors" }
func (kolorsEngine) ext() string           { return "png" }
func (kolorsEngine) needCensorSkip(error) bool { return false } // Kolors 无硬审核跳过

func (kolorsEngine) genOnce(prompt, dstPath string) error {
	if kolorsAPIKey == "" {
		return fmt.Errorf("未设置 KOLORS_API_KEY 环境变量")
	}
	body, err := json.Marshal(kolorsReq{
		Model:         kolorsModel,
		Prompt:        prompt,
		ImageSize:     kolorsSize,
		NumSteps:      kolorsSteps,
		GuidanceScale: kolorsGuidance,
	})
	if err != nil {
		return fmt.Errorf("编码请求失败: %w", err)
	}
	req, err := http.NewRequest("POST", kolorsEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+kolorsAPIKey)

	client := &http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var out kolorsResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return fmt.Errorf("解析响应失败: %w, body=%s", err, string(raw))
	}
	if out.Code != 0 && out.Code != 200 && out.Message != "" {
		return fmt.Errorf("API 错误 code=%d: %s", out.Code, out.Message)
	}
	if len(out.Images) == 0 || out.Images[0].URL == "" {
		return fmt.Errorf("响应无图片 URL, body=%s", string(raw))
	}
	return downloadTo(out.Images[0].URL, dstPath, client)
}

// -----------------------------------------------------------------------------
// 公共下载 + 重试
// -----------------------------------------------------------------------------

func downloadTo(url, dstPath string, client *http.Client) error {
	imgResp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("下载图片失败: %w", err)
	}
	defer imgResp.Body.Close()
	if imgResp.StatusCode != 200 {
		return fmt.Errorf("下载图片 HTTP %d", imgResp.StatusCode)
	}
	imgBytes, err := io.ReadAll(imgResp.Body)
	if err != nil {
		return fmt.Errorf("读取图片流失败: %w", err)
	}
	if len(imgBytes) == 0 {
		return fmt.Errorf("图片为空")
	}
	if err := os.WriteFile(dstPath, imgBytes, 0644); err != nil {
		return fmt.Errorf("写文件失败: %w", err)
	}
	return nil
}

// generateImage 带重试的图片生成：无限重试，指数退避 2s→60s。
// 内容审核类错误（如 CogView 的 1301）直接返回，跳过该幕。
func generateImage(eng imageEngine, sceneID int, prompt, dstPath string) error {
	const maxWait = 60 * time.Second
	wait := 2 * time.Second
	for attempt := 1; ; attempt++ {
		err := eng.genOnce(prompt, dstPath)
		if err == nil {
			return nil
		}
		if eng.needCensorSkip(err) {
			return err
		}
		fmt.Printf("  [幕 %d 出图第 %d 次失败，%v 后重试] %v\n", sceneID, attempt, wait, err)
		time.Sleep(wait)
		wait *= 2
		if wait > maxWait {
			wait = maxWait
		}
	}
}

// -----------------------------------------------------------------------------
// buildImgPromptPrompt：让 LLM 阅读一幕正文，输出可直接喂给文生图模型的画面提示词。
//
// 两种引擎的提示词策略不同：
//   - CogView：内容审核严，需要洗白（禁止血/尸/杀/西式地名等）
//   - Kolors：审核宽松，允许更真实的剧情描写与光影氛围
// -----------------------------------------------------------------------------

func buildImgPromptPrompt(text string, engineName string) string {
	if engineName == "Kolors" {
		return buildKolorsPrompt(text)
	}
	return buildCogViewPrompt(text)
}

func buildKolorsPrompt(text string) string {
	return `你是一名文生图 prompt 工程师，为一款以明末崇祯朝为背景的历史向文字冒险游戏生成 16:9 电影感画面提示词。

**核心风格锚点（必须写进提示词开头）**：明代中国风古画，绢本水墨工笔重彩描金，电影感史诗构图，画面精致细腻，光影层次分明。

要求：
1. 阅读一幕剧情正文，抓住其"环境+时刻+氛围"三要素，扩写成一段生动的画面描述；
2. **画面必须无人、无文字招牌**——只描绘场景、建筑、天地、光影、烟云、器物；剧情里的人物/动作只可通过环境痕迹暗示（如空椅、遗落的兵器、烧毁的城墙、无人的祭坛等）；
3. **可以真实描写氛围**：血色残阳、篝火冲天、雪原冰河、断壁残垣、狼藉战场、烟尘弥漫、白绫垂落、烛火摇曳、月落乌啼、香烟袅袅——只要不出现具体人物即可，务必贴合剧情张力；
4. 剧情发生在日本/南洋/美洲/西域/云南等异域时，允许描写当地地理特征（樱花寺、竹楼、南洋古港、密林、雪山、雪原、部落石城、西班牙式石堡废墟等），但整体画风保持"东方古画质感"，避免出现现代建筑、汽车、电线、玻璃幕墙、招牌文字、路灯、水泥这些破坏时代感的元素；
5. **强调光影和氛围**：晨曦金光、血色黄昏、月夜幽蓝、烛火摇曳、雪雾弥漫、烟尘漫天、瑞霭紫气、雷云低垂 —— 光影是塑造电影感的关键；
6. 长度 100-180 字，中文，用逗号顿号分隔的关键词式描述，不要分行不要标号不要引号不要"提示词："前缀，直接输出画面；
7. 开头必须以"明代中国风古画，绢本水墨工笔重彩描金"起头。

一幕剧情正文：
"""
` + text + `
"""

请输出画面提示词：`
}

func buildCogViewPrompt(text string) string {
	return `你是一名文生图 prompt 工程师。请阅读以下一幕游戏剧情正文，输出一段"仅描述古代画面场景"的中文提示词，供后续文生图模型使用。

**核心风格锚点（必须写进提示词开头）**：明代中国风古画、水墨工笔、绢本设色、无人物、无文字。

严格遵守：
1. 只描述**古代场景、古建筑、山川自然、物件、光影、天气、氛围**，禁止出现任何人物、动作、情节、对话、情绪；
2. 严禁出现以下词汇：皇帝、崇祯、朱由检、紫禁城、太子、血、尸、杀、死、斩、首、刀剑、战、屠、贼、贵妃、AI、语言模型；如需表达皇宫用"古代大型宫殿"或"皇家宫苑"；
3. **严禁出现任何暗示人群/仪式的词汇**：朝会、百官、大典、仪仗、跪拜、群臣、宴会、审讯、议事、朝拜、卫兵、士兵、宫女、太监；
4. **严禁出现任何暗示灾难/暴力/死亡的画面词汇**：白绫、绳套、丝帛、槐树、煤山、火光、火焰、浓烟、黑烟、烟尘、烽火、烽烟、焦烟、灰烬、末日、颓败、残垣、废墟、断壁、尸、血；
5. **严禁出现任何模糊现代/易被误画为现代的词汇**：街道、街景、街市、市集、码头、港口、商铺、店铺、招牌、灯箱、路灯、水泥、大厦、装置、涂鸦、市场；
6. **严禁出现任何西方/异域元素**：西班牙、教堂、总督府、别墅、洋楼、欧式、大理石、雕像、罗马、白人、原住民、部落、祭坛、圆顶、尖顶；
7. **严禁出现任何"抽象/艺术/现代摄影"风格词**：抽象、艺术装置、彩色纹理、几何图案、特写、微距、俯视、航拍、街拍、写实照片、彩色数字、荧光、霓虹；
8. **只能画中国明代场景**——即使剧情发生在日本/南洋/美洲/西域/台湾/云南，也用"东方古代山川/古寺/古亭/古船/山寨/竹楼"这种通用东方古风描述替代；
9. 画面必须**空景无人无字**；
10. 输出一整段中文，60-120 字，用逗号/顿号分隔关键词式描述，不要分行不要标号不要引号；开头必须用"明代中国风古画，水墨工笔"起头；
11. 直接输出结果，不要解释。

一幕剧情正文：
"""
` + text + `
"""

请输出画面提示词：`
}

// polishImgPrompt 清理 LLM 输出，追加固定风格后缀。
func polishImgPrompt(raw, engineName string) string {
	s := strings.TrimSpace(raw)
	for _, p := range []string{"提示词：", "提示词:", "画面提示词：", "画面提示词:", "输出：", "输出:", "\"", "'", "“", "”"} {
		s = strings.TrimPrefix(s, p)
	}
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "\n\n"); idx > 0 {
		s = s[:idx]
	}
	s = strings.ReplaceAll(s, "\n", "，")
	s = strings.TrimSuffix(s, "。")

	if engineName == "Kolors" {
		return s + "。明代中国风古画，绢本水墨工笔重彩描金，电影感构图，无人物无文字，古代东方场景，禁止现代建筑/汽车/电线/玻璃楼/招牌文字/现代摄影"
	}
	return s + "。明代中国风古画，绢本水墨工笔重彩，电影感构图，空无一人无文字，古代东方山水建筑，禁止任何现代建筑/汽车/电线/玻璃楼/招牌/文字/西式建筑/人物/雕塑/剪影/抽象艺术"
}

// -----------------------------------------------------------------------------
// 图片生成主流程
// -----------------------------------------------------------------------------

func runImgMode(llm LLM, maxScenes int, eng imageEngine) {
	entries, err := os.ReadDir(ScenesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取 %s 失败: %v\n", ScenesDir, err)
		os.Exit(1)
	}

	var allIDs []int
	for _, e := range entries {
		if !e.IsDir() || !sceneNum.MatchString(e.Name()) {
			continue
		}
		id, _ := strconv.Atoi(e.Name())
		allIDs = append(allIDs, id)
	}
	sort.Ints(allIDs)

	var todo []int
	skip := 0
	for _, id := range allIDs {
		dir := filepath.Join(ScenesDir, strconv.Itoa(id))
		if hasBg(dir) {
			skip++
			continue
		}
		todo = append(todo, id)
		if len(todo) >= maxScenes {
			break
		}
	}
	fmt.Printf("[img/%s] 待处理 %d 幕（跳过已有 %d 幕），串行处理\n", eng.name(), len(todo), skip)

	imgOK, censored := 0, 0
	var censoredIDs []int
	start := time.Now()

	for i, id := range todo {
		dir := filepath.Join(ScenesDir, strconv.Itoa(id))
		imgTxt := filepath.Join(dir, "img.txt")

		fmt.Printf("[开始 %d/%d] 幕 %d\n", i+1, len(todo), id)

		var prompt string
		if fi, err := os.Stat(imgTxt); err == nil && fi.Size() > 0 {
			b, _ := os.ReadFile(imgTxt)
			prompt = strings.TrimSpace(string(b))
			fmt.Printf("  [幕 %d] 复用现有 img.txt\n", id)
		} else {
			textBytes, err := os.ReadFile(filepath.Join(dir, "text.txt"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "  [幕 %d 跳过] 读 text.txt 失败: %v\n", id, err)
				continue
			}
			text := strings.TrimSpace(string(textBytes))
			if text == "" {
				fmt.Fprintf(os.Stderr, "  [幕 %d 跳过] text.txt 为空\n", id)
				continue
			}
			fmt.Printf("  [幕 %d] 调 LLM 生成提示词 ...\n", id)
			raw, err := llm.Complete(buildImgPromptPrompt(text, eng.name()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "  [幕 %d 跳过] LLM 最终失败: %v\n", id, err)
				continue
			}
			prompt = polishImgPrompt(raw, eng.name())
			if err := os.WriteFile(imgTxt, []byte(prompt), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "  [幕 %d 跳过] 写 img.txt 失败: %v\n", id, err)
				continue
			}
			fmt.Printf("  [幕 %d] 提示词已写入 img.txt\n", id)
		}

		dst := filepath.Join(dir, "bg."+eng.ext())
		fmt.Printf("  [幕 %d] 调 %s 出图 ...\n", id, eng.name())
		if err := generateImage(eng, id, prompt, dst); err != nil {
			if eng.needCensorSkip(err) {
				censored++
				censoredIDs = append(censoredIDs, id)
				fmt.Fprintf(os.Stderr, "  [幕 %d 跳过] 内容审核拦截 (id=%d): %v\n", id, id, err)
				continue
			}
			fmt.Fprintf(os.Stderr, "  [幕 %d 跳过] 出图最终失败: %v\n", id, err)
			continue
		}
		imgOK++
		remain := len(todo) - i - 1
		elapsed := time.Since(start)
		speed := float64(imgOK) / elapsed.Seconds()
		etaStr := "?"
		if speed > 0 {
			etaStr = fmtDuration(float64(remain) / speed)
		}
		fmt.Printf("[完成 %d/%d] 幕 %d | 拦截 %d | 速度 %.2f 幕/秒 | ETA %s\n",
			imgOK, len(todo), id, censored, speed, etaStr)
	}

	fmt.Printf("\n完成。新出图 %d 张，审核拦截 %d 幕（id=%v），跳过（已有）%d，耗时 %v。\n",
		imgOK, censored, censoredIDs, skip, time.Since(start).Round(time.Second))
}

// hasBg 检查是否已经存在任一支持扩展名的背景图。
func hasBg(dir string) bool {
	for _, ext := range []string{"png", "jpg", "jpeg", "webp"} {
		if _, err := os.Stat(filepath.Join(dir, "bg."+ext)); err == nil {
			return true
		}
	}
	return false
}
