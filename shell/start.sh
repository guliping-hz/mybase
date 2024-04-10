#!/bin/bash

echo "use exeNameReplace"
#增加执行权限
chmod +xxx "$(pwd)/exeNameReplace"

#exe后面带上 & 防止关闭终端，就把go进程结束掉;而且必须以shell脚本的形式启动 & 才能起作用。
#直接在终端中敲下面的命令关掉终端，进程还是结束了。。

#捕获崩溃异常，dlv
ulimit -c unlimited
export GOTRACEBACK=crash

#多个服务只是目录的不同
#必须重定向到/dev/null，否则远程启动时，关闭终端，进程也会结束
sudo $(pwd)/exeNameReplace &
#sudo $(pwd)/exeNameReplace > /dev/null &
