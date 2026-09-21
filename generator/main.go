package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ContentDir  = "../content"
	ScenesDir   = "../content/scenes"
	MetaFile    = "../content/meta.txt"
	OutlineFile = "../content/outline.txt"
)

// -----------------------------------------------------------------------------
// 树模型：以 outline.txt 每行为一条"根→叶"的路径，共享前缀 = 同一个节点。
// 每个 text 节点分配一个唯一 scene ID（1 = 根）。
// -----------------------------------------------------------------------------

type node struct {
	id       int      // scene id
	seed     string   // 大纲里这一节点的原文（"text..." 里 "text" 之后的那段）
	depth    int      // 根为 0
	parent   *node    // 根为 nil
	pchoice  string   // 从父到本节点所选的 choice 原文（"choice..." 里 "choice" 之后的那段），根为 ""
	children []*edge  // 从本节点出去的 choice 边（按大纲首次出现顺序）
}

type edge struct {
	choice string // choice 原文
	to     *node
}

// -----------------------------------------------------------------------------
// outline 解析
// -----------------------------------------------------------------------------

func parseOutline(path string) *node {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("读取 %s 失败: %v", path, err)
	}
	defer f.Close()

	var root *node
	nextID := 0
	// key = 从根到该 text 节点的"seed 序列"拼接，用于去重
	nodeKey := func(seeds []string) string { return strings.Join(seeds, "\x1f") }
	byKey := map[string]*node{}

	newNode := func(seed string, depth int, parent *node, pchoice string) *node {
		nextID++
		return &node{id: nextID, seed: seed, depth: depth, parent: parent, pchoice: pchoice}
	}

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, " - ")
		if len(parts)%2 == 0 {
			log.Fatalf("outline 行格式错误（token 数必须为奇数）: %s", line)
		}
		// 沿路径 walk / create
		var cur *node
		var seeds []string
		for i, tok := range parts {
			tok = strings.TrimSpace(tok)
			if i%2 == 0 {
				// text 节点
				if !strings.HasPrefix(tok, "text") {
					log.Fatalf("第 %d 个 token 应为 text: %s", i, line)
				}
				seed := strings.TrimSpace(strings.TrimPrefix(tok, "text"))
				seeds = append(seeds, seed)
				key := nodeKey(seeds)
				if existing, ok := byKey[key]; ok {
					cur = existing
				} else {
					if cur == nil {
						// 根
						nd := newNode(seed, 0, nil, "")
						byKey[key] = nd
						if root == nil {
							root = nd
						} else if root.seed != seed {
							log.Fatalf("outline 有多个不同的根 text: %q vs %q", root.seed, seed)
						} else {
							// 同一根，理论上 byKey 已命中，这里不该到；保底
							nd = root
							byKey[key] = nd
						}
						cur = nd
					} else {
						log.Fatalf("内部错误：非根节点未在 byKey 中命中: %s", key)
					}
				}
			} else {
				// choice 边
				if !strings.HasPrefix(tok, "choice") {
					log.Fatalf("第 %d 个 token 应为 choice: %s", i, line)
				}
				choice := strings.TrimSpace(strings.TrimPrefix(tok, "choice"))
				// 下一 token 是子 text
				if i+1 >= len(parts) {
					log.Fatalf("choice 后缺少 text: %s", line)
				}
				childTok := strings.TrimSpace(parts[i+1])
				if !strings.HasPrefix(childTok, "text") {
					log.Fatalf("choice 后应为 text: %s", line)
				}
				childSeed := strings.TrimSpace(strings.TrimPrefix(childTok, "text"))

				// 查找/创建子节点
				childSeeds := append([]string{}, seeds...)
				childSeeds = append(childSeeds, childSeed)
				ckey := nodeKey(childSeeds)

				var child *node
				if existing, ok := byKey[ckey]; ok {
					child = existing
				} else {
					child = newNode(childSeed, cur.depth+1, cur, choice)
					byKey[ckey] = child
					cur.children = append(cur.children, &edge{choice: choice, to: child})
				}

				// 若已存在，还要检查这条 choice 边在父节点上是否已建立
				found := false
				for _, e := range cur.children {
					if e.to == child {
						found = true
						// 校验 choice 文本一致（同一子节点必定通过同一 choice 到达）
						if e.choice != choice {
							log.Fatalf("同一子节点被两种 choice 到达: %q vs %q -> %s", e.choice, choice, childSeed)
						}
						break
					}
				}
				if !found {
					cur.children = append(cur.children, &edge{choice: choice, to: child})
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("扫描 outline 失败: %v", err)
	}
	if root == nil {
		log.Fatalf("outline 为空")
	}
	return root
}

// flatten 按 BFS 顺序返回所有节点，保证父先于子。
func flatten(root *node) []*node {
	var out []*node
	q := []*node{root}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		out = append(out, n)
		for _, e := range n.children {
			q = append(q, e.to)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].id < out[j].id })
	return out
}

// ancestors 返回从根到该节点（含）的节点序列。
func ancestors(n *node) []*node {
	var chain []*node
	for x := n; x != nil; x = x.parent {
		chain = append([]*node{x}, chain...)
	}
	return chain
}

// -----------------------------------------------------------------------------
// LLM prompt
// -----------------------------------------------------------------------------

const styleRules = `【写作约束（必须遵守）】
- 第二人称"你"（你 = 崇祯 = 玩家）
- 正文严格控制在 130-180 字之间
- 以白话为主，可用少量文言点缀（旨意、称谓、奏对），避免通篇文白夹杂
- 画面感优先：光线、气味、声音、身体触感、他人神态
- 忠实于给定的"本幕剧情要点"——要点里发生的事就是既成事实，必须落实到正文里，不可推翻、不可回避、不可用心理活动代替
- 不要给正文加标题、编号、旁白、场景标签
- 不要在正文里替玩家做决定，不要写"你可以选择……"这类元话语
- 只输出剧情正文本身，不加任何解释、开场白、Markdown、引号包裹`

func buildRootPrompt(world, seed string) string {
	return fmt.Sprintf(`你是一位擅长历史向互动小说的写作者，现在要为一款文字冒险游戏写开场（第 1 幕）。

【世界观】
%s

【本幕剧情要点（大纲原文，需忠实展开）】
%s

【任务】把这段"剧情要点"扩写成一段有画面感的开场正文。
要求聚焦"此刻"的感官冲击（光线/气味/声音/身体触感），点出"崇祯十六年冬、李自成东进、清军窥伺"的背景，写出"两个头脑并存"的错乱感——现代大脑正在急速处理"我穿越了"这个事实，而身体本能地扮演一个疲惫多疑的中年皇帝。结尾停在一个自然的抉择点上，但不要写出选项本身。

%s`, world, seed, styleRules)
}

func buildChildPrompt(world string, chain []*node, cur *node, isEnding bool) string {
	// 拼接祖先路径
	var b strings.Builder
	for i, n := range chain {
		if n == cur {
			break
		}
		fmt.Fprintf(&b, "第 %d 幕正文要点：%s\n", i+1, n.seed)
		// n 与下一个 chain 元素之间的 choice = 下一个节点的 pchoice
		if i+1 < len(chain) {
			fmt.Fprintf(&b, "→ 玩家选择：%s\n\n", chain[i+1].pchoice)
		}
	}

	endingHint := ""
	tail := "结尾停在一个新的、具体的困境上，让玩家感觉'该做点什么了'，但不要写出选项。"
	if isEnding {
		endingHint = "\n【本幕是本条故事线的最终结局】"
		tail = "结尾必须收束——局面已定、大势已去（或已成），去掉所有问句和抉择抛问，写出'故事已终'的沉重感或释然。允许时间快进（'三日后'、'半年后'），但必须落回一个具体的、带感官细节的场景收尾。"
	}

	return fmt.Sprintf(`你是一位擅长历史向互动小说的写作者，正在续写一款文字冒险游戏的分支剧情。

【世界观】
%s

【至此为止的故事脉络（既成事实，不可推翻）】
%s【玩家刚刚做出的选择（就是主角实际采取的行动）】
%s

【本幕剧情要点（大纲原文，需忠实展开——里头说了什么就得发生什么，不要用"你在犹豫"这种心理描写代替情节）】
%s%s

【任务】把这段"剧情要点"扩写成本幕的完整正文。
从"主角刚做出这个选择的瞬间或紧接着的后果"切入，把选择的直接影响与要点里的事件演出来（他人反应、事态变化、身体感受）。%s

%s`, world, b.String(), cur.pchoice, cur.seed, endingHint, tail, styleRules)
}

// -----------------------------------------------------------------------------
// 主流程
// -----------------------------------------------------------------------------

func main() {
	maxScenes := flag.Int("max", 100000, "最大处理幕数（含跳过），默认不限")
	provider := flag.String("provider", "longcat", "LLM 提供方：mimo | longcat | mock")
	imageMode := flag.Bool("image", false, "图片模式：对每幕先用 LLM 生成 img.txt 再调生图 API 出 bg 图")
	imgEngine := flag.String("imgengine", "kolors", "生图引擎：kolors（SiliconFlow Kwai-Kolors，2016x1120 PNG）| cogview（智谱 CogView-3-Flash，1344x768 JPG）")
	flag.Parse()

	world := mustRead("world.txt")

	var llm LLM
	switch *provider {
	case "mimo":
		llm = NewMimoLLM()
	case "longcat":
		llm = NewLongCatLLM("")
	case "mock":
		llm = NewMockLLM()
	default:
		log.Fatalf("未知 provider: %s", *provider)
	}
	fmt.Printf("使用 %s\n", *provider)

	if *imageMode {
		var eng imageEngine
		switch *imgEngine {
		case "kolors":
			eng = kolorsEngine{}
		case "cogview":
			eng = cogViewEngine{}
		default:
			log.Fatalf("未知 imgengine: %s", *imgEngine)
		}
		runImgMode(llm, *maxScenes, eng)
		return
	}

	if err := os.MkdirAll(ScenesDir, 0755); err != nil {
		panic(err)
	}
	ensureMeta()

	root := parseOutline(OutlineFile)
	nodes := flatten(root)
	total := len(nodes)
	limit := *maxScenes
	if limit > total {
		limit = total
	}

	// 预扫描：统计本次真正需要 LLM 生成的幕数（有 text.txt 的直接跳过）
	toGen := 0
	for i, n := range nodes {
		if i >= limit {
			break
		}
		if _, err := os.Stat(filepath.Join(ScenesDir, strconv.Itoa(n.id), "text.txt")); err != nil {
			toGen++
		}
	}
	remaining := toGen
	fmt.Printf("outline 共 %d 幕，本次将处理 %d 幕（需生成 %d 幕）\n", total, limit, toGen)

	processed, generated, skipped := 0, 0, 0
	startTime := time.Now()
	for _, n := range nodes {
		if processed >= limit {
			break
		}
		processed++
		dir := filepath.Join(ScenesDir, strconv.Itoa(n.id))
		os.MkdirAll(dir, 0755)

		// 先写 choices.txt（无 LLM 依赖，可反复覆盖以保证与 outline 同步）
		writeChoicesFile(dir, n)

		textPath := filepath.Join(dir, "text.txt")
		if _, err := os.Stat(textPath); err == nil {
			skipped++
			fmt.Printf("[跳过] 幕 %d（父=%s, 深=%d）已存在\n", n.id, parentIDStr(n), n.depth)
			continue
		}

		isEnding := len(n.children) == 0
		var prompt string
		if n.parent == nil {
			prompt = buildRootPrompt(world, n.seed)
		} else {
			chain := ancestors(n)
			prompt = buildChildPrompt(world, chain, n, isEnding)
		}

		// 生成 + 校验，失败重试（LLM 层已有网络层重试）
		for attempt := 1; ; attempt++ {
			raw, err := llm.Complete(prompt)
			if err != nil {
				log.Fatalf("LLM 调用失败: %v", err)
			}
			text := cleanLLMText(raw)
			if err := os.WriteFile(textPath, []byte(text), 0644); err != nil {
				log.Fatalf("写 text.txt 失败: %v", err)
			}
			if ok, reason := validateSceneText(dir); ok {
				break
			} else {
				fmt.Fprintf(os.Stderr, "  [幕 %d 第 %d 次校验失败：%s，重试]\n", n.id, attempt, reason)
				os.Remove(textPath)
			}
		}
		generated++
		remaining--
		tag := "生成"
		if isEnding {
			tag = "生成[结局]"
		}
		elapsed := time.Since(startTime)
		speed := float64(generated) / elapsed.Seconds()
		etaStr := "?"
		if speed > 0 {
			etaSec := float64(remaining) / speed
			etaStr = fmtDuration(etaSec)
		}
		fmt.Printf("[%s] 幕 %d（父=%s, 深=%d, 子=%d）| 剩余 %d | 速度 %.2f 幕/秒 | ETA %s\n",
			tag, n.id, parentIDStr(n), n.depth, len(n.children), remaining, speed, etaStr)
	}

	fmt.Printf("\n完成。共处理 %d 幕（新生成 %d，跳过 %d）\n", processed, generated, skipped)
}

func parentIDStr(n *node) string {
	if n.parent == nil {
		return "-"
	}
	return strconv.Itoa(n.parent.id)
}

// writeChoicesFile 将本节点的 choices.txt 写好。
// 叶子节点（结局）写入 "结局"（前端识别后显示重新开始按钮）。
func writeChoicesFile(dir string, n *node) {
	cp := filepath.Join(dir, "choices.txt")
	if len(n.children) == 0 {
		os.WriteFile(cp, []byte("结局"), 0644)
		return
	}
	var lines []string
	for _, e := range n.children {
		lines = append(lines, fmt.Sprintf("%s | %d", e.choice, e.to.id))
	}
	os.WriteFile(cp, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// -----------------------------------------------------------------------------
// 工具
// -----------------------------------------------------------------------------

func ensureMeta() {
	if _, err := os.Stat(MetaFile); err == nil {
		return
	}
	os.WriteFile(MetaFile, []byte("title=天命·崇祯十六年\nstart=1\ntypeSpeed=40\n"), 0644)
}

func mustRead(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		panic(fmt.Errorf("读取 %s 失败: %w", p, err))
	}
	return string(b)
}

func fmtDuration(sec float64) string {
	if sec < 60 {
		return fmt.Sprintf("%ds", int(sec))
	}
	if sec < 3600 {
		return fmt.Sprintf("%dm%ds", int(sec)/60, int(sec)%60)
	}
	h := int(sec) / 3600
	m := (int(sec) % 3600) / 60
	return fmt.Sprintf("%dh%dm", h, m)
}
