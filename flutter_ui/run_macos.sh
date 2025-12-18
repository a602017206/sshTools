#!/bin/bash

# Flutter macOS启动脚本
# 这个脚本设置正确的Ruby环境并启动Flutter应用

cd "$(dirname "$0")"

echo "==> 加载Ruby环境..."
# 加载RVM并使用Ruby 3.3.4
source ~/.rvm/scripts/rvm
rvm use 3.3.4 2>/dev/null

# 设置正确的PATH
export PATH="$HOME/.rvm/gems/ruby-3.3.4/bin:$HOME/.rvm/rubies/ruby-3.3.4/bin:$PATH"

echo "==> 启动Flutter应用..."
echo "    后端API: http://localhost:8080"
echo "    确保Go后端服务器正在运行"
echo ""

# 运行Flutter应用
/Users/dingwei/development/flutter/bin/flutter run -d macos
