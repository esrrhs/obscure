# obscure

纯前端文字冒险：HTML + CSS + JS，无构建、无依赖。

每一幕 = 背景图 + 打字机对白（可叠立绘）。点击推进对白与下一幕，直到结局。

## 在线试玩

👉 **[https://esrrhs.github.io/obscure/](https://esrrhs.github.io/obscure/)**

![cover](cover.jpg)

## 本地运行

```bash
python3 -m http.server 8000
# 浏览器打开 http://localhost:8000/
```

> `file://` 协议因 CORS 限制无法加载资源，需通过静态服务器访问。

## 目录结构

```
/
├── index.html              # 入口，跳转到 framework/
├── framework/              # 游戏框架（读取并呈现 content/）
│   ├── index.html
│   ├── css/style.css
│   └── js/game.js
├── content/                # 剧本与素材
│   ├── meta.txt            # 标题、起始幕、打字速度
│   ├── style.txt           # 视觉风格说明
│   ├── outline.txt         # 剧情大纲
│   ├── characters/<名>/    # 每个角色目录
│   │   ├── name.txt        # 显示名
│   │   └── portrait.png    # 立绘（透明底 PNG）
│   └── scenes/<数字ID>/    # 每一幕
│       ├── text.txt        # 正文
│       ├── choices.txt     # 选项 → 下一幕
│       └── bg.png          # 背景图
├── cover.jpg               # 封面图
└── logo.png                # Logo
```

框架只负责「读取 → 表现」；替换一套符合格式的 `content/` 即可换剧本。

## 内容格式

### `content/meta.txt`

```
title=obscure
start=1
typeSpeed=40
```

### `content/characters/<名>/`

| 文件 | 说明 |
|------|------|
| `name.txt` | 角色显示名，用于匹配对白说话人 |
| `portrait.png` | 立绘，建议透明底 PNG；缺失则该角色说话时不显示立绘 |

### `content/scenes/<ID>/`

| 文件 | 说明 |
|------|------|
| `text.txt` | 正文，保留换行；格式：`说话人："台词"` 或纯旁白 |
| `choices.txt` | 跳转：线性写 `继续 \| 下一幕ID`（点击直接进入）；多选项时显示按钮；`结局` 或文件缺失 = 结局幕 |
| `bg.png` | 背景图，支持 png / jpg / webp 等格式 |

## 表现

- 背景铺满，切换时交叉淡入淡出
- 有对白时左侧显示对应角色立绘；旁白时自动收起
- 对话按句推进：一句打字机出现，点击进入下一句；打字中再点可跳过当前句
- 一幕说完后：线性剧情点一下进入下一幕；多分支显示选项按钮；结局幕显示「重新开始」

## 角色

| 目录 | 角色名 | 备注 |
|------|--------|------|
| `me` | 我（主角） | 第一人称，无立绘 |
| `eileen` | 艾琳 | 女主 |
| `laozhu` | 老朱 / 猪妖 | |
| `suan` | 算命先生 | |
| `pm` | PM | |
| `luren` | 路人 | |
| `sun` | 孙大圣 | |
| `zhu` | 猪八戒 | |
| `sha` | 沙僧 | |
| `tang` | 唐僧 | |
