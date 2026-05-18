#!/bin/bash
# 本地构建前端资源脚本
# Local frontend build script

set -e

# 添加 bun 到 PATH
export PATH="$HOME/.bun/bin:$PATH"

echo "==> 开始构建前端资源 (Building frontend assets)..."

# 读取版本号
VERSION=$(cat VERSION)
export VITE_REACT_APP_VERSION=$VERSION

# 构建 default 主题
echo "==> 构建 default 主题 (Building default theme)..."
cd web/default
bun install
DISABLE_ESLINT_PLUGIN='true' bun run build
cd ../..

# 构建 classic 主题
echo "==> 构建 classic 主题 (Building classic theme)..."
cd web/classic
bun install
NODE_OPTIONS="--max-old-space-size=4096" bun run build
cd ../..

echo "==> 前端构建完成 (Frontend build completed)"
echo "==> 现在可以使用 Dockerfile.prebuilt 构建 Docker 镜像"
echo "==> Now you can build Docker image with Dockerfile.prebuilt"
