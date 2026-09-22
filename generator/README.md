# 图片生成

默认使用**硅基流动 Kolors**（与 TheMandate 相同），输出约 **2016×1120**。

## 配置

在项目根目录创建 `.env`（已 gitignore，**不要提交**）：

```
KOLORS_API_KEY=sk-xxx
```

可选：

```
IMAGE_ENGINE=kolors          # 默认 kolors；也可 freellmapi
FREELLMAPI_API_KEY=...       # 仅 IMAGE_ENGINE=freellmapi 时需要
```

## 用法

```bash
cd generator
python3 gen_images.py bg       # 仅背景（缺什么补什么）
python3 gen_images.py cover    # 仅封面
python3 gen_images.py char     # 仅立绘
python3 gen_images.py all      # 全部
python3 gen_images.py bg 5     # 最多新生成 5 张
```

已有 `bg.*` / `portrait.*` 会跳过。提示词在各目录 `img.txt`。
