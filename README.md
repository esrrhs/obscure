# obscure

纯前端文字冒险（HTML + CSS + JS，无构建、无依赖）。
每一幕 = 背景图 + 打字机文字 + 选项；点选项跳转下一幕，直到结局。

架构同 [TheMandate](https://github.com/esrrhs/TheMandate)：框架与内容解耦，`content/` 按约定格式即可跑。

## 本地运行

```bash
python3 -m http.server 8000
# 打开 http://localhost:8000/ 或 http://localhost:8000/framework/
```

## 目录

```
/
├── index.html          # 入口 → framework/
├── framework/          # 表现层
│   ├── index.html
│   ├── css/style.css
│   └── js/game.js
├── content/            # 剧本与资源
│   ├── meta.txt        # 标题、起始幕、打字速度
│   ├── style.txt       # 视觉风格说明
│   ├── outline.txt     # 剧情大纲
│   ├── characters/     # 立绘（portrait + img.txt）
│   └── scenes/<id>/    # text.txt / choices.txt / bg.* / img.txt
├── generator/          # 生图脚本
├── cover.jpg
└── cover_prompt.txt
```

## 内容格式

- `meta.txt`：`title=` / `start=` / `typeSpeed=`
- 每幕 `text.txt` 为正文；`choices.txt` 每行 `选项 | 下一幕ID`，内容为 `结局` 则为本局结束
- `bg.png` / `bg.jpg` 等为背景；`img.txt` 为文生图提示词（框架忽略）

## 风格

中国水墨奇幻古风，绢本水墨工笔重彩描金，电影感构图，暗色调水墨氤氲。
详见 `content/style.txt`。

## 生图

见 `generator/README.md`。
