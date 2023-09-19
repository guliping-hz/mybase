#!python3.11
# author:guliping

import os
import sys
import importlib
import subprocess
import math
import time
import inspect
import http.client
import hashlib
import urllib.request, ssl, http.cookiejar
import json


def my_print(*args, end: str | None = None, nofile: bool = False):
    # 获取上一行调用的位置信息
    caller_frame = inspect.currentframe().f_back
    caller_info = inspect.getframeinfo(caller_frame)
    if nofile:
        print(*args, end=end)
    else:
        print(f"{caller_info.filename}:{caller_info.lineno}", *args, end=end)


def check_and_install_dependency(p_import: str, package: str | None = None):
    try:
        if not package:
            package = p_import
        importlib.import_module(p_import)
    except Exception as e:
        print("check_and_install_dependency e=", e)
        print(f"{package} is not installed. Installing...")
        try:
            subprocess.check_call(["pip", "install", package])
            print(f"{package} has been installed.")
        except Exception as e:
            print(f"install {package} fail,e=", e)
            sys.exit(1)


# 检查并安装缺少的依赖项
check_and_install_dependency("paramiko")
check_and_install_dependency("scp")
from scp import SCPClient
import paramiko


def build_go(
    srcDir: str, targetDir: str, targetName: str, platform: str = "linux"
) -> bool:
    """
    @platform=linux|windows|darwin
    """
    ret = False
    try:
        my_print("compile...")
        os.environ["GOARCH"] = "amd64"
        os.environ["GOOS"] = platform
        os.environ["GOTRACEBACK"] = "all"
        subprocess.check_output(
            ["go", "build", "-o", f"{targetDir}/{targetName}", "-C", srcDir]
        )
        my_print(f"Go program compiled successfully to {targetDir}/{targetName}")
        ret = True
    except subprocess.CalledProcessError as e:
        my_print("Error compiling Go program:", e)
    finally:
        del os.environ["GOARCH"]
        del os.environ["GOOS"]
        del os.environ["GOTRACEBACK"]
        return ret


def remote_exec(ip: str, command: str, port: int = 22, user: str = "root") -> bool:
    # 创建SSH客户端
    ssh_client = paramiko.SSHClient()
    ssh_client.load_system_host_keys()
    ssh_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    ret = False
    try:
        # 连接到远程服务器
        ssh_client.connect(ip, port, user)

        # 在远程服务器上执行命令
        stdin, stdout, stderr = ssh_client.exec_command(command)

        # 打印命令输出
        my_print(command, "=>")

        while True:
            out1 = stdout.read().decode()
            err1 = stderr.read().decode()
            if out1 != "":
                my_print(out1, end="", nofile=True)
            elif err1 != "":
                my_print(err1, end="", nofile=True)
            else:
                break
        my_print(nofile=True)
        # # 使用scp模块上传文件到远程服务器
        # with SCPClient(ssh_client.get_transport()) as scp:
        #     scp.put("local_file.txt", "remote_file.txt")  # 上传文件

        #     # 或者从远程服务器下载文件到本地
        #     # scp.get("remote_file.txt", "local_file.txt")  # 下载文件
        ret = True
    except Exception as e:
        my_print("Error remote_exec", command, ":", e)
    finally:
        # 关闭SSH连接
        ssh_client.close()
        return ret


def remote_put(
    ip: str,
    src: str,
    dest: str,
    isDir: bool = False,
    port: int = 22,
    user: str = "root",
) -> bool:
    # 创建SSH客户端
    ssh_client = paramiko.SSHClient()
    ssh_client.load_system_host_keys()
    ssh_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    ret = False
    try:
        # 连接到远程服务器
        ssh_client.connect(ip, port, user)
        # 确保目录存在
        destDir = os.path.dirname(dest)
        stdin, stdout, stderr = ssh_client.exec_command(f"mkdir -p {destDir}")
        out1 = stdout.read().decode()
        err1 = stderr.read().decode()
        if out1 != "":
            my_print(out1)
        if err1 != "":
            my_print(err1)

        # 自定义进度回调函数
        curFileName = None

        def progress(filename, size, sent):
            nonlocal curFileName
            percent = sent / size * 100
            figure = "*" * math.floor(percent / 5)
            if filename == curFileName:
                my_print(f"\rUploading {filename} {figure} {percent:.2f}%", end="")
            else:
                curFileName = filename
                my_print(f"\nUploading {filename} {figure} {percent:.2f}%", end="")

        # # 使用scp模块上传文件到远程服务器
        with SCPClient(ssh_client.get_transport(), progress=progress) as scp:
            scp.put(src, dest, recursive=isDir, preserve_times=True)  # 上传文件

            # 或者从远程服务器下载文件到本地
            # scp.get("remote_file.txt", "local_file.txt")  # 下载文件
        my_print("\nscp put", src, "=>", dest)
        ret = True
    except Exception as e:
        my_print("Error remote_scp", src, "=>", dest, ":", e)
    finally:
        # 关闭SSH连接
        ssh_client.close()
        return ret


def md5_str(str):
    return hashlib.md5(bytes(str, "utf_8")).hexdigest()


def http_with_cookie(url: str, param: str | dict, timeout: int = 1800) -> str | None:
    """
    #@param url 被请求的URL地址(必需)
    #@param param POST参数，可以是字符串或字典(必需)
    #@param timeout 超时时间默认1800秒
    #@return str
    """
    # cookie
    parsedUrl = urllib.parse.urlparse(url)
    cookieFile = "./" + md5_str(parsedUrl.netloc) + ".cookie"
    cookieObj = http.cookiejar.MozillaCookieJar(cookieFile)
    if os.path.exists(cookieFile):
        cookieObj.load(cookieFile, ignore_discard=True, ignore_expires=True)
    cookieHandler = urllib.request.HTTPCookieProcessor(cookieObj)

    # https 忽略证书校验
    httpsHandler = urllib.request.HTTPSHandler(context=ssl._create_unverified_context())

    # post 表单
    data = urllib.parse.urlencode(param).encode("utf-8")

    try:
        # 请求
        req = urllib.request.Request(url, data)
        opener = urllib.request.build_opener(cookieHandler, httpsHandler)
        response = opener.open(req, timeout=timeout)
        my_print(url, "=>", response.status, response.reason)

        # 保存cookie
        cookieObj.save(ignore_discard=True, ignore_expires=True)

        result = response.read()
        my_print("=>", result)
        if type(result) == bytes:
            result = result.decode("utf-8")
        return result
    except Exception as e:
        my_print(url, "fail =>", e)
        return None


class BtApi:
    __BT_KEY = "2VF0LUAr395MFdewjuivQKCMHFGIgD0ZZDwKXY0B5g7wo8"
    __BT_PANEL = "http://ip:8888"

    # 如果希望多台面板，可以在实例化对象时，将面板地址与密钥传入
    def __init__(self, bt_panel=None, bt_key=None):
        if bt_panel:
            self.__BT_PANEL = bt_panel
            self.__BT_KEY = bt_key

    # 取面板日志
    def get_logs(self):
        # 拼接URL地址
        url = self.__BT_PANEL + "/data?action=getData"

        # 准备POST数据
        param = self.__get_key_data()  # 取签名
        param["table"] = "logs"
        param["limit"] = 10
        param["tojs"] = "test"

        # 请求面板接口
        result = http_with_cookie(url, param, 1800)

        # 解析JSON数据
        return json.loads(result)

    # 系统状态
    def get_system_total(self):
        url = self.__BT_PANEL + "/system?action=GetSystemTotal"
        param = self.__get_key_data()  # 取签名
        result = http_with_cookie(url, param, 1800)
        if result:
            return json.loads(result)
        return None

    # 建站
    def add_site(
        self, domain: str, version: str = "00", path: str | None = None, port: int = 80
    ):
        url = self.__BT_PANEL + "/site?action=AddSite"
        param = self.__get_key_data()  # 取签名
        chromeCatchParam = {
            "webname": '{"domain": "' + domain + '", "domainlist": [], "count": 0}',
            "type": "PHP",
            "port": str(port),
            "ps": domain,
            "path": path or f"/www/wwwroot/{domain}",
            "type_id": "0",
            "version": version,
            "ftp": "false",
            "sql": "false",
            "codeing": "utf8mb4",
        }
        param.update(chromeCatchParam)
        result = http_with_cookie(url, param, 1800)
        if result:
            jsonResult = json.loads(result)
            if jsonResult["siteStatus"]:
                my_print(f"{domain} 建站成功")
                return jsonResult
        my_print(f"{domain} 建站失败")
        return None

    # 设置网站根路径
    def set_path(self, id: int, path: str):
        url = self.__BT_PANEL + "/site?action=SetPath"
        param = self.__get_key_data()  # 取签名
        param["id"] = str(id)
        param["path"] = path
        result = http_with_cookie(url, param, 1800)
        if result:
            jsonResult = json.loads(result)
            if jsonResult["siteStatus"]:
                my_print(f"{path} 更新成功")
                return jsonResult
        my_print(f"{path} 更新成功")
        return None

    # 获取网站配置文件
    def get_file_body(self, domain: str):
        url = self.__BT_PANEL + "/files?action=GetFileBody"
        param = self.__get_key_data()  # 取签名
        param["path"] = f"/www/server/panel/vhost/nginx/{domain}.conf"
        result = http_with_cookie(url, param, 1800)
        if result:
            return json.loads(result)
        return None

    # 设置网站配置文件
    def save_file_body(self, domain: str, body: str, path: str):
        url = self.__BT_PANEL + "/files?action=SaveFileBody"
        param = self.__get_key_data()  # 取签名
        chromeCatchParam = {
            "data": body,
            "path": path,
            "encoding": "utf-8",
        }
        param.update(chromeCatchParam)
        result = http_with_cookie(url, param, 1800)
        my_print("save_file_body", result)
        if result:
            jsonResult = json.loads(result)
            if jsonResult["status"]:
                my_print(f"{domain} 更新成功")
                return
        my_print(f"{domain} 更新失败")

    # 设置网站伪静态配置
    def save_file_body_pretend_static(self, domain: str, body: str):
        self.save_file_body(
            domain, body, f"/www/server/panel/vhost/rewrite/{domain}.conf"
        )

    # 设置网站nginx配置
    def save_file_body_nginx(self, domain: str, body: str):
        self.save_file_body(
            domain, body, f"/www/server/panel/vhost/nginx/{domain}.conf"
        )

    # 构造带有签名的关联数组
    def __get_key_data(self):
        now_time = int(time.time())
        param = {
            "request_token": md5_str(str(now_time) + "" + md5_str(self.__BT_KEY)),
            "request_time": now_time,
        }
        return param


class BaseSetup:
    def __init__(self, ip: str, port: int = 22, user: str = "root") -> None:
        self.ip = ip
        self.port = port
        self.user = user

    def remote_exec(self, commond: str):
        return remote_exec(self.ip, commond, self.port, self.user)

    def remote_put(self, src: str, dest: str, isDir: bool = False):
        return remote_put(self.ip, src, dest, isDir, self.port, self.user)

    def get_acme(hostName: str):
        return f"""
"/root/.acme.sh/"acme.sh --install-cert -d {hostName} \
        --key-file       /root/cert/{hostName}/key.pem  \
        --fullchain-file /root/cert/{hostName}/cert.pem \
        --reloadcmd     "nginx -s reload"
"""


class Setup(BaseSetup):
    def __init__(self, args) -> None:
        super().__init__(args.ip, args.port, args.user)

        self.dir = args.dir
        self.exe = args.exe
        self.screen = args.screen
        self.buildDir = args.build_dir
        self.outDir = args.out_dir
        self.shellOn = args.shell
        self.cacheDirNames = args.cache_dir_names
        self.envs = args.envs
        self.crontab = args.crontab

        screenName = self.exe + self.screen

        self.shell_start = f"""
#!/bin/bash

echo "use {self.exe}"
#增加执行权限
chmod +xxx "$(pwd)/{self.exe}"

#exe后面带上 & 防止关闭终端，就把go进程结束掉;而且必须以shell脚本的形式启动 & 才能起作用。
#直接在终端中敲下面的命令关掉终端，进程还是结束了。。

#捕获崩溃异常，dlv
ulimit -c unlimited
export GOTRACEBACK=crash

#多个服务只是目录的不同
#必须重定向到/dev/null，否则远程启动时，关闭终端，进程也会结束
$(pwd)/{self.exe} &
#$(pwd)/{self.exe} > /dev/null &
"""
        self.shell_end = """
#!/bin/bash

#通用停止当前目录启动的进程
#kill -9 发送 SIGKILL信号
#ps -aux | grep "$(pwd)" | grep -v "grep" | awk '{print $2}' | xargs kill -9
#15 发送 SIGTERM信号，允许程序优雅退出
# ps -aux | grep "$(pwd)" | grep -v "grep" | awk '{print $2}' | xargs kill -15

#同一个目录有多个exe的时候，适用于下面的结束
function stop(){
    pid=$(ps -aux | grep "$(pwd)/$1" | grep -v "grep" | awk '{print $2}')
    if [ -n "$pid" ]; then
        echo "stop $1 pid: $pid"
        echo "$pid" | xargs kill "$2"
    else
        echo "No process:$(pwd)/$1 to kill"
    fi
}

function stop15(){
    stop "$1" -15
}

stop15 {self.exe}
"""

        self.shell_restart = """
#!/bin/bash

#上传后，如果运行报错。更改回车符：goland->File->File Properties->Line Separators->LF-Unix & macOs(\n)

#通用重启当前目录的进程。
chmod +xxx $(pwd)/start.sh
chmod +xxx $(pwd)/end.sh

$(pwd)/end.sh
echo "sleep 2s..."
sleep 2s
$(pwd)/start.sh
"""
        self.shell_screen = (
            """
#退出对应的screen  ||true 忽略执行错误
screen -S """
            + screenName
            + """ -X quit || true
#重新创建一个新的screen
screen -dmS """
            + screenName
            + """ || false

#指定执行脚本内容 
script="cd """
            + self.dir
            + """  && ./start.sh"
#离屏执行一段内容
screen -S """
            + screenName
            + """ -X eval "screen" "-X" "stuff '${script} \n'"
"""
        )

    def start(self) -> bool:
        # 编译代码
        if not build_go(self.buildDir, self.outDir, self.exe):
            return False

        # 创建服务器目录环境
        if not self.remote_exec(f"mkdir -p {self.dir}"):
            return False
        for i in range(len(self.cacheDirNames)):
            subDir = self.dir + "/" + self.cacheDirNames[i]
            if not self.remote_exec(f"mkdir -p {subDir}"):
                return False

        # 重命名
        self.remote_exec(f"cd {self.dir} && rm {self.exe}.bak")
        self.remote_exec(f"cd {self.dir} && mv {self.exe} {self.exe}.bak")

        # 上传文件
        exeFullPath = self.outDir + "/" + self.exe
        if not self.remote_put(exeFullPath, self.dir):  # + "/" + self.exe
            return False

        if self.shellOn:
            # 上传
            if not self.remote_exec(f"echo {self.shell_end} > {self.dir}/end.sh"):
                return False
            if not self.remote_exec(f"echo {self.shell_start} > {self.dir}/start.sh"):
                return False
            if not self.remote_exec(
                f"echo {self.shell_restart} > {self.dir}/restart.sh"
            ):
                return False
            if not self.remote_exec(f"echo {self.shell_screen} > {self.dir}/screen.sh"):
                return False

        for i in range(0, len(self.envs), 3):
            here, there, isDir = self.envs[i], self.envs[i + 1], self.envs[i + 2] == "1"
            if not self.remote_put(here, self.dir + "/" + there, isDir):
                return False

        if self.shellOn:
            # 关闭
            if not self.remote_exec(f"cd {self.dir} && chmod +xxx ./restart.sh"):
                return False
            if not self.remote_exec(
                f"cd {self.dir} && chmod +xxx ./end.sh && ./end.sh"
            ):
                return False

            # sleep 3秒
            my_print("sleep 3...")
            time.sleep(3)

            # 启动
            if not self.remote_exec(f"cd {self.dir} && chmod +xxx ./start.sh"):
                return False
            if not self.remote_exec(
                f"cd {self.dir} && chmod +xxx ./screen.sh && ./screen.sh"
            ):
                return False

        # 定时任务配置
        if self.crontab != "":
            crontabPath = self.dir + "/" + self.exe + ".crontab"
            if not self.remote_put(self.crontab, crontabPath):
                return False
            if not self.remote_exec(f"crontab {crontabPath}"):
                return False

        my_print(f"setup finished {self.ip}:{self.dir}/{self.exe}")


def parse_to_setup():
    import argparse

    # 创建 ArgumentParser 对象
    parser = argparse.ArgumentParser(description="自动化部署")

    # 添加命令行参数
    parser.add_argument("-ip", required=True, help="远程IP")
    parser.add_argument("-port", default=22, type=int, help="远程端口")
    parser.add_argument("-exe", required=True, help="可执行程序名")
    parser.add_argument("-screen", default="", help="screen名后缀,防止冲突")
    parser.add_argument("-dir", required=True, help="远程目录")
    parser.add_argument("-build_dir", required=True, help="编译目录")
    parser.add_argument("-out_dir", required=True, help="编译输出目录")
    parser.add_argument("-user", default="root", help="远程用户")

    # 这个True False的只能用两个减
    parser.add_argument(
        "--shell", action=argparse.BooleanOptionalAction, help="是否上传通用shell脚本"
    )
    parser.add_argument("-cache_dir_names", nargs="*", help="远程备用目录名,以dir指定的目录为当前前缀")
    parser.add_argument("-envs", nargs="*", help="环境配置文件。例:本地文件名,服务器文件名,1目录/0文件")
    parser.add_argument("-crontab", default="", help="远程用户")

    args = parser.parse_args()
    my_print("parseArgs=", args)
    setup = Setup(args)
    setup.start()


def test():
    # remote_exec("127.0.0.1", "ls -l")
    # remote_exec("127.0.0.1", "pwd")
    cwd = os.getcwd()  # 获取当前目录
    # remote_put("127.0.0.1", cwd + "/../auth/cfg/yqw2.json", "/opt/auth/auth.json")

    pass


if __name__ == "__main__":
    # test()
    # print(__file__, os.path.dirname(__file__))
    parse_to_setup()

    # print(BaseSetup.get_acme("1.cn"))
