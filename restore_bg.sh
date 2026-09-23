#!/bin/bash
# 从 9ee6669 提取原始 Ren'Py 背景图，按 source_bg.txt 映射写入 content/scenes/*/bg.jpg(png)
set -e

COMMIT="9ee6669"

for i in $(seq 1 27); do
  src_name=$(cat "content/scenes/$i/source_bg.txt" 2>/dev/null | tr -d '\r\n')
  if [ -z "$src_name" ]; then
    echo "scene $i: 无 source_bg.txt，跳过"
    continue
  fi

  # Ren'Py 文件名带空格，如 "bg 000100"
  src_file="game/obscure/game/images/${src_name}.png"
  out_file="content/scenes/$i/bg.png"

  echo -n "scene $i ($src_name) ... "

  # 从 git 提取文件
  if git cat-file -e "${COMMIT}:${src_file}" 2>/dev/null; then
    git show "${COMMIT}:${src_file}" > "$out_file"
    # 删除旧的 jpg
    rm -f "content/scenes/$i/bg.jpg"
    echo "OK -> bg.png"
  else
    echo "MISSING in commit!"
  fi
done

echo ""
echo "=== 完成 ==="
