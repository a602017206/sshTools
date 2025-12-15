#!/bin/bash

# 快速修复客户端按钮点击问题的脚本

echo "🔧 修复 Wails 客户端按钮点击问题..."
echo ""

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ 错误：请在项目根目录运行此脚本"
    exit 1
fi

echo "📝 验证修复是否已应用..."
if grep -q "webkit-app-region: no-drag" frontend/src/components/ConnectionManager.svelte; then
    echo "✅ 修复已应用到 ConnectionManager.svelte"
else
    echo "⚠️  警告：ConnectionManager.svelte 中未找到修复"
fi

echo ""
echo "🔄 重新构建应用..."
go build -v

if [ $? -eq 0 ]; then
    echo "✅ Go 后端构建成功"
else
    echo "❌ Go 后端构建失败"
    exit 1
fi

echo ""
echo "📦 前端准备..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "📥 安装前端依赖..."
    npm install
fi

echo ""
echo "✅ 修复完成！"
echo ""
echo "🚀 启动应用："
echo "   $HOME/go/bin/wails dev"
echo ""
echo "📋 测试清单："
echo "   1. 点击 '+ 新建连接' 按钮"
echo "   2. 在表单中点击各个按钮（保存、取消、测试连接）"
echo "   3. 在连接列表中点击 '连接'、'编辑'、'删除' 按钮"
echo ""
echo "💡 如果按钮仍无法点击，请查看 DEBUG_CLIENT_BUTTONS.md 获取更多调试方法"
