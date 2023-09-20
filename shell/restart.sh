#!/bin/bash

#上传后，如果运行报错。更改回车符：goland->File->File Properties->Line Separators->LF-Unix & macOs

#通用重启当前目录的进程。
chmod +xxx $(pwd)/start.sh
chmod +xxx $(pwd)/end.sh

$(pwd)/end.sh
echo "sleep 2s..."
sleep 2s
$(pwd)/start.sh
