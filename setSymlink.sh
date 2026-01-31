#!/bin/bash

# 要建立 symlink 的目錄名稱清單
MODULES=(common lang logger utils)

echo "🔁 建立 shared-modules 的 symlink 到本專案目錄..."

for dir in "${MODULES[@]}"; do
  # 如果資料夾或 symlink 存在，先刪除
  if [ -e "$dir" ] || [ -L "$dir" ]; then
    echo "🧹 移除已存在的 $dir"
    rm -rf "$dir"
  fi

  # 建立 symlink
  ln -s ../shared-modules/"$dir" ./"$dir"
  echo "✅ 建立 symlink: $dir → ../shared-modules/$dir"
done

echo "🎉 所有 symlink 已建立完成。"

