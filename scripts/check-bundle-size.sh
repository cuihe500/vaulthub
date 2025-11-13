#!/bin/bash
# Bundle 大小检查脚本
# 用法: ./check-bundle-size.sh [MAX_MAIN_KB] [MAX_CHUNK_KB]
# 示例: ./check-bundle-size.sh 50 350

set -e

MAX_MAIN_BUNDLE_SIZE=${1:-50}
MAX_CHUNK_SIZE=${2:-350}

echo "检查前端bundle大小..."
echo "主入口阈值: ${MAX_MAIN_BUNDLE_SIZE}KB (gzip)"
echo "Chunk阈值: ${MAX_CHUNK_SIZE}KB (gzip)"
echo ""

cd web/dist/assets/js

# 检查主入口bundle
check_main_bundle() {
  local main_bundle=$(find . -name "index-*.js" -type f | head -1)

  if [ -n "$main_bundle" ]; then
    local main_size=$(gzip -c "$main_bundle" | wc -c)
    local main_size_kb=$((main_size / 1024))
    echo "主入口bundle: ${main_size_kb}KB (gzip)"

    if [ "$main_size_kb" -gt "$MAX_MAIN_BUNDLE_SIZE" ]; then
      echo "❌ 主入口bundle过大! 超过${MAX_MAIN_BUNDLE_SIZE}KB阈值"
      exit 1
    fi
  fi
}

# 检查chunks
check_chunks() {
  local has_chunks=false

  for file in chunk-*.js; do
    # 检查文件是否存在（避免glob不匹配时的字面量）
    [ ! -f "$file" ] && continue

    has_chunks=true
    local size=$(gzip -c "$file" | wc -c)
    local size_kb=$((size / 1024))
    local filename=$(basename "$file")
    echo "  ${filename}: ${size_kb}KB (gzip)"

    if [ "$size_kb" -gt "$MAX_CHUNK_SIZE" ]; then
      echo "❌ chunk过大! ${filename} 超过${MAX_CHUNK_SIZE}KB阈值"
      exit 1
    fi
  done

  if [ "$has_chunks" = false ]; then
    echo "  (未找到chunk文件)"
  fi
}

check_main_bundle
check_chunks

echo ""
echo "✅ Bundle大小检查通过"
