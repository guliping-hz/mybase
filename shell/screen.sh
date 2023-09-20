#退出对应的screen  ||true 忽略执行错误
screen -S screenNameReplace -X quit || true
#重新创建一个新的screen
screen -dmS screenNameReplace || false

#指定执行脚本内容
script="cd linuxDirReplace  && ./start.sh"
#离屏执行一段内容
screen -S screenNameReplace -X eval "screen" "-X" "stuff '${script} \n'"
