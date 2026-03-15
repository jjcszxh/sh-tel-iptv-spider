#!/bin/bash
set -e

# 项目名称
PROJECT_NAME="sh-tel-iptv-spider"

# 打印版本信息（可选）
echo "Building $PROJECT_NAME, version: $VERSION"

# 创建 build 目录（防止不存在报错）
mkdir -p build

# 使用 gox 进行交叉编译，指定要编译的操作系统和架构
# 针对 Linux 64 位, Linux 32 位, Windows 64 位, Windows 32 位, Linux ARM 32 位, Linux ARM 64 位
CGO_ENABLED=0 gox -ldflags "-s -w ${LDFLAGS}" \
  -output="build/${PROJECT_NAME}_{{.OS}}_{{.Arch}}" \
  --osarch="linux/amd64 linux/386 windows/amd64 windows/386 linux/arm linux/arm64"

# 进入 build 目录
cd build

# 生成多种校验和
# 生成标准校验和
rhash -r -a . -o checksums
# 生成 BSD 校验和
rhash -r -a --bsd . -o checksums-bsd
# 生成包含哈希顺序的校验和列表
rhash --list-hashes > checksums_hashes_order

# 打印完成信息
echo "Build and checksum generation completed!"
