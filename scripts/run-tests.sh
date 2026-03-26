#!/bin/bash
# run-tests.sh - 单元测试门禁脚本
# 用法: ./scripts/run-tests.sh

set -e

echo "Running unit tests with race detection..."
go test -v -race ./...

echo "All tests passed!"
