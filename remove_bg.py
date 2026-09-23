#!/usr/bin/env python3
"""批量去背景脚本，用 rembg Python API 处理所有 portrait.jpg"""
import glob
import os
from pathlib import Path
from PIL import Image
from rembg import remove

portraits = sorted(glob.glob("content/characters/*/portrait.jpg"))
print(f"找到 {len(portraits)} 个立绘文件\n")

for jpg_path in portraits:
    char = Path(jpg_path).parent.name
    out_path = str(Path(jpg_path).parent / "portrait.png")
    print(f"[{char}] 处理中 ...", end=" ", flush=True)
    with open(jpg_path, "rb") as f:
        input_data = f.read()
    output_data = remove(input_data)
    with open(out_path, "wb") as f:
        f.write(output_data)
    # 校验
    img = Image.open(out_path)
    mode = img.mode
    size = img.size
    has_alpha = mode == "RGBA"
    # 检查是否真的有透明像素
    if has_alpha:
        alpha = img.split()[3]
        transparent_pixels = sum(1 for p in alpha.getdata() if p < 128)
        total = size[0] * size[1]
        pct = transparent_pixels / total * 100
        print(f"OK ({mode}, {size[0]}x{size[1]}, 透明像素占 {pct:.1f}%)")
    else:
        print(f"WARNING: 无 alpha 通道 ({mode})")

print("\n=== 全部完成 ===")
