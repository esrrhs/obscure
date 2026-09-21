# obscure

纯前端文字冒险：HTML + CSS + JS，无构建、无依赖。

每一幕 = 背景图 + 打字机文字 + 若干选项。点选项进入下一幕，直到结局。

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
| `choices.txt` | 每行 `选项文字 \| 下一幕ID`；写 `结局` 或文件缺失 = 结局幕 |
| `bg.*` | 背景图，支持 png / jpg / webp 等 |
| `img.txt` | 生图用提示词，框架忽略 |

## 表现

- 背景铺满，切换时交叉淡入淡出
- 文字打字机效果；点击可跳过
- 打完后选项依次浮现
- 结局幕显示「重新开始」

## 视觉风格

中国水墨奇幻古风，绢本水墨工笔重彩描金，电影感构图，暗色调水墨氤氲。详见 `content/style.txt`。

## 生图

```bash
cd generator
export FREELLMAPI_API_KEY=freellmapi-xxx
python3 gen_images.py          # 背景 + 立绘 + 封面
python3 gen_images.py bg       # 仅背景
python3 gen_images.py char     # 仅立绘
```

已有图片会跳过，可续跑。更多说明见 `generator/README.md`。
