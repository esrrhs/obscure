# 图片生成

为 `content/scenes/` 与 `content/characters/` 生成背景与立绘。

## 用法

通过 FreeLLMAPI 调 NVIDIA FLUX：

```bash
export FREELLMAPI_API_KEY=freellmapi-xxx
export FREELLMAPI_URL=http://127.0.0.1:13001/v1/images/generations   # 可选
export IMAGE_MODEL=black-forest-labs/flux.1-dev                      # 可选

python3 gen_images.py          # 背景 + 立绘 + 封面
python3 gen_images.py bg       # 仅背景
python3 gen_images.py char     # 仅立绘
python3 gen_images.py cover    # 仅封面
python3 gen_images.py bg 5     # 最多新生成 5 张
```

已有 `bg.*` / `portrait.*` 会跳过，可安全续跑。提示词在各目录 `img.txt`。

## 风格

见 `../content/style.txt`。提示词以英文书写（FLUX 遵从度更好），锚定水墨工笔奇幻。
