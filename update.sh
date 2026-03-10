#!/bin/bash

cd /vol1/1000/web/sharednodes

python3 /vol1/1000/web/sharednodes/filter_nodes.py
echo "✅ 文章生成完毕！"

python3 /vol1/1000/web/sharednodes/git_update.py
echo "✅ README生成完毕！"

git add .
git commit -m '节点更新'
git push
echo "✅ 节点更新完成！"

cd /vol1/1000/web/hexo_blog
hexo g
hexo d
echo "✅ 博客更新完成！"