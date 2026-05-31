#!/bin/bash

# 1. 进入你的私有项目目录 (请修改为你真实的绝对路径)
cd /home/ian/vscode/Go/gobook || exit

# 2. 获取当前时间作为提交信息
TIME=$(date "+%Y-%m-%d %H:%M:%S")

# 3. 将所有变更添加到 Git
git add .

# 4. 提交更改 (如果没有任何改动，commit 会失败，但没关系，脚本会继续)
git commit -m "Daily auto backup: $TIME"

# 5. 推送到 GitHub (假设你的主分支叫 main)
git push origin main