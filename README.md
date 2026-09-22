# obscure

纯前端文字冒险：HTML + CSS + JS，无构建、无依赖。

每一幕 = 背景图 + 打字机文字（可叠立绘）。点击推进对白与下一幕，直到结局。

## 在线试玩

👉 **[https://esrrhs.github.io/obscure/](https://esrrhs.github.io/obscure/)**

![cover](cover.jpg)

## 本地运行

```bash
python3 -m http.server 8000
# 浏览器打开 http://localhost:8000/
```

`file://` 会因 CORS 无法加载资源，需要静态服务器。

## 目录结构

```
/
├── index.html              # 入口，跳转到 framework/
├── framework/              # 游戏框架（读取并表现 content/）
│   ├── index.html
│   ├── css/style.css
│   └── js/game.js
├── content/                # 剧本与素材
│   ├── meta.txt            # 标题、起始幕、打字速度
│   ├── style.txt           # 视觉风格说明
│   ├── outline.txt         # 剧情大纲
│   ├── characters/<名>/    # 立绘 portrait.* + img.txt
│   └── scenes/<数字ID>/    # 每一幕
│       ├── text.txt        # 正文
│       ├── choices.txt     # 选项 → 下一幕
│       ├── bg.jpg          # 背景图
│       └── img.txt         # 文生图提示词
├── generator/              # 批量生图脚本
├── cover.jpg
└── cover_prompt.txt
```

框架只负责「读取 → 表现」；换一套符合格式的 `content/` 即可换剧本。

## 内容格式

### `content/meta.txt`

```
title=obscure
start=1
typeSpeed=40
```

### `content/scenes/<ID>/`

| 文件 | 说明 |
|------|------|
| `text.txt` | 正文，保留换行 |
| `choices.txt` | 跳转：线性写 `继续 \| 下一幕ID`（界面不显示按钮，点一下即进下一幕）；多选项时显示按钮；`结局` 或缺失 = 结局幕 |
| `bg.*` | 背景图，支持 png / jpg / webp 等 |
| `img.txt` | 生图用提示词，框架忽略 |

## 表现

- 背景铺满，切换时交叉淡入淡出
- 有对白时左侧显示对应角色立绘；旁白时收起
- 对话按句推进：一句打字机出现，点击进入下一句；打字中再点可跳过当前句
- 一幕说完后：线性剧情再点一下进入下一幕；多分支才显示选项；结局幕显示「重新开始」

## 视觉风格

玄幻恐怖国风：血月黑雾、腥红渍迹、腐朽妖气，电影感暗色调。详见 `content/style.txt`。

## 生图

默认 **硅基流动 Kolors**（2016×1120）。在项目根 `.env` 写入 `KOLORS_API_KEY`（已 gitignore，勿提交）。

```bash
cd generator
python3 gen_images.py bg       # 背景
python3 gen_images.py char     # 立绘
```

更多说明见 `generator/README.md`。
