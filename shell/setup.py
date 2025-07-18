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
import select
import socket
from zipfile import ZipFile


def change_line_end_from_win_to_linux(file: str):
    with open(file, "rb") as f:
        content = f.read()
        # print(content)
        contentNew = content.decode("utf8").replace("\r\n", "\n")
        # print(contentNew.encode("utf8"))

        with open(file, "wb") as f2:
            f2.write(contentNew.encode("utf8"))
            f2.flush()


def get_dir_files(dir: str, ignores: dict[str, bool] | None = None) -> list[str]:
    files = []
    allChildrenFiles = []
    for dirpath, dirnames, filenames in os.walk(dir):
        print("dirpath=", dirpath)
        dirpath = dirpath.replace("\\", "/")
        # print("dirnames=", dirnames)
        # print("filenames=", filenames)

        nowDir = os.path.basename(dirpath)
        if ignores:
            if nowDir in ignores:
                print("skip")
                continue
            find = False
            for k in ignores.keys():
                if dirpath.find("/" + k + "/") != -1:
                    find = True
                    break
            if find:
                print("skip")
                continue

        for file in filenames:
            files.append(file)
            fullpath = dirpath + "/" + file  # os.path.join(dirpath, file)

            # 更新行尾结束符
            if fullpath.endswith(".sh"):
                change_line_end_from_win_to_linux(fullpath)

            # print("fullpath=", fullpath)
            allChildrenFiles.append(fullpath)
    # return files
    return allChildrenFiles


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
        print(f"{package} is already installed.")
    except ImportError:
        print(f"{package} is not installed. Installing...")
        try:
            # 升级pip
            subprocess.check_call(
                ["python3", "-m", "pip", "install", "--upgrade", "pip"]
            )
            subprocess.check_call(["pip3", "install", package])
            print(f"{package} has been installed.")
        except subprocess.CalledProcessError:
            print(f"install {package} fail,exit...")
            sys.exit(1)


# 检查并安装缺少的依赖项
check_and_install_dependency("paramiko")
check_and_install_dependency("scp")
check_and_install_dependency("socks", "PySocks")
from scp import SCPClient
import paramiko
import socks


def build_go(
    srcDir: str | None,
    srcFile: str | None,
    targetDir: str,
    targetName: str,
    buildvcs: bool = False,
    platform: str = "linux",
) -> bool:
    """
    @srcDir 指定编译目录
    @srcFile 指定编译文件
    @platform=linux|windows|darwin
    """
    my_print(f"Go program to compile dir={srcDir} file={srcFile}")
    ret = False
    try:
        my_print("compile...")
        os.environ["GOARCH"] = "amd64"
        os.environ["GOOS"] = platform
        os.environ["GOTRACEBACK"] = "all"

        if srcDir and srcDir != "":
            # my_print("srcDir")
            commond = ["go", "build", "-C", srcDir]
            if not buildvcs:
                commond.append("-buildvcs=false")
            commond.extend(
                [
                    "-o",
                    f"{targetDir}/{targetName}",
                ]
            )

            my_print(commond)
            subprocess.check_output(commond)
            my_print(f"Go program compiled successfully to {targetDir}/{targetName}")
            ret = True
        elif srcFile and srcFile != "":
            # my_print("srcFile")
            srcDir = os.path.dirname(srcFile)
            srcFileBase = os.path.basename(srcFile)

            commond = ["go", "build", "-C", srcDir]
            if not buildvcs:
                commond.append("-buildvcs=false")
            commond.extend(
                [
                    "-o",
                    f"{targetDir}/{targetName}",
                    srcFileBase,
                ]
            )
            my_print(commond)
            subprocess.check_output(commond)
            my_print(f"Go program compiled successfully to {targetDir}/{targetName}")
            ret = True
        else:
            my_print(f"Go program compiled failed not set build dir or file")
            ret = False
    except subprocess.CalledProcessError as e:
        my_print("subprocess.CalledProcessError compiling Go program:", e)
    except Exception as e:
        my_print("Exception compiling Go program:", e)
    finally:
        my_print(f"end compiling Go program: {ret}")
        if "GOARCH" in os.environ:
            del os.environ["GOARCH"]
        if "GOOS" in os.environ:
            del os.environ["GOOS"]
        if "GOTRACEBACK" in os.environ:
            del os.environ["GOTRACEBACK"]

        if not ret:
            os._exit(1)
        return ret


def remote_exec(
    ip: str,
    command: str,
    port: int = 22,
    user: str = "root",
    socksHost: str = "localhost",
    socksPort: int = 0,
    raiseError: bool = True,
) -> bool:
    sock = None
    if socksPort:
        # print(f"use socks {socksHost}:{socksPort}")
        # socks.set_default_proxy(socks.SOCKS5, socksHost, socksPort)
        sock = socks.socksocket()
        sock.set_proxy(socks.SOCKS5, socksHost, socksPort, False)
        sock.connect((ip, port))

    # 创建SSH客户端
    ssh_client = paramiko.SSHClient()
    ssh_client.load_system_host_keys()
    ssh_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    ret = False
    try:
        # 连接到远程服务器
        ssh_client.connect(ip, port, user, sock=sock)

        # 在远程服务器上执行命令
        stdin, stdout, stderr = ssh_client.exec_command(command)

        # 打印命令输出
        my_print(command, "=>")

        # 循环读取输出直到输出流关闭
        while not stdout.channel.exit_status_ready():
            # 读取标准输出
            output = stdout.readline()
            if output:
                print(output.strip())
        # 读取剩余的标准输出
        remaining_output = stdout.read().decode()
        if remaining_output:
            print(remaining_output.strip())

        # # 读取标准错误
        error = stderr.read().decode()
        if error:
            my_print("Error:", error, end="", nofile=True)
        my_print(nofile=True)

        # # 使用scp模块上传文件到远程服务器
        # with SCPClient(ssh_client.get_transport()) as scp:
        #     scp.put("local_file.txt", "remote_file.txt")  # 上传文件

        #     # 或者从远程服务器下载文件到本地
        #     # scp.get("remote_file.txt", "local_file.txt")  # 下载文件
        ret = True
    except Exception as e:
        my_print("Error remote_exec:", command, ":", e, ret)
    finally:
        # 关闭SSH连接
        ssh_client.close()

        if not ret and raiseError:
            raise RuntimeError()
        return ret


def remote_put(
    ip: str,
    src: str,
    dest: str,
    isDir: bool = False,
    port: int = 22,
    user: str = "root",
    socksHost: str = "localhost",
    socksPort: int = 0,
    raiseError: bool = True,
) -> bool:
    """
    :param socksHost 代理host
    :param socksPort 代理端口
    """

    # 优先把脚本换成linux行尾符号
    if not isDir and src.endswith(".sh"):
        change_line_end_from_win_to_linux(src)

    sock = None
    if socksPort:  # 过滤 None 和 0
        print(f"use socks {socksHost}:{socksPort}")
        # socks.set_default_proxy(socks.SOCKS5, socksHost, socksPort)
        sock = socks.socksocket()
        sock.set_proxy(socks.SOCKS5, socksHost, socksPort, False)
        sock.connect((ip, port))

    # 创建SSH客户端
    ssh_client = paramiko.SSHClient()
    ssh_client.load_system_host_keys()
    ssh_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    ret = False
    try:
        # 连接到远程服务器
        ssh_client.connect(ip, port, user, sock=sock)
        # 确保目录存在
        destDir = os.path.dirname(dest)
        stdin, stdout, stderr = ssh_client.exec_command(f"sudo mkdir -p {destDir}")
        out1 = stdout.read().decode()
        err1 = stderr.read().decode()
        if out1 != "":
            my_print(out1)
        if err1 != "":
            my_print(err1)

        # 自定义进度回调函数
        curFileName = None

        my_print("\nscp put", src, "=>", dest)

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
        print()
        ret = True
    except Exception as e:
        my_print("Error remote_scp", src, "=>", dest, ":", e)
    finally:
        # 关闭SSH连接
        ssh_client.close()

        if not ret and raiseError:
            raise RuntimeError()
        return ret


def md5_str(str) -> str:
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

    # 取面板网站
    def get_sites(self):
        # 拼接URL地址
        url = self.__BT_PANEL + "/data?action=getData"

        # 准备POST数据
        param = self.__get_key_data()  # 取签名
        param["table"] = "sites"
        param["limit"] = ""
        param["p"] = "1"
        param["search"] = ""
        param["type"] = "0"

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
            if jsonResult and "siteStatus" in jsonResult and jsonResult["siteStatus"]:
                my_print(f"{domain} 建站成功")
                return jsonResult
        my_print(f"{domain} 建站失败", result)
        return None

    # 申请SSL证书
    def apply_cert_api(
        self,
        id: int,
        autoTo: int,
        domains: list[str],
    ):
        url = self.__BT_PANEL + "/acme?action=apply_cert_api"
        param = self.__get_key_data()  # 取签名
        param["domains"] = json.dumps(domains)
        param["auth_type"] = "http"
        param["auth_to"] = str(autoTo)
        param["auto_wildcard"] = str(0)
        param["id"] = str(id)
        result = http_with_cookie(url, param, 1800)
        if result:
            jsonResult = json.loads(result)
            if jsonResult["siteStatus"]:
                my_print(f"{domains} 证书申请成功")
                return jsonResult
        my_print(f"{domains} 证书申请失败")
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
    def __init__(
        self,
        ip: str,
        port: int = 22,
        user: str = "root",
        vpn_host: str = "localhost",
        vpn_port: int = 0,
    ) -> None:
        self.ip = ip
        self.port = port
        self.user = user

        self.socksHost = vpn_host
        self.socksPort = vpn_port

        if self.socksPort:
            print(f"use socks {self.socksHost}:{self.socksPort}")

    def remote_exec(self, commond: str) -> bool:
        return remote_exec(self.ip, commond, self.port, self.user)

    def remote_put(self, src: str, dest: str, isDir: bool = False) -> bool:
        return remote_put(
            self.ip,
            src,
            dest,
            isDir,
            self.port,
            self.user,
            self.socksHost,
            self.socksPort,
        )

    def get_acme(hostName: str) -> str:
        return f"""
/root/.acme.sh/acme.sh --install-cert -d {hostName} \
--key-file /root/cert/{hostName}/key.pem  \
--fullchain-file /root/cert/{hostName}/cert.pem \
--reloadcmd "nginx -s reload"
"""

    def get_start() -> str:
        return """#!/bin/bash

echo "use exeNameReplace"
#增加执行权限
sudo chmod +xxx "$(pwd)/exeNameReplace"

#exe后面带上 & 防止关闭终端，就把go进程结束掉;而且必须以shell脚本的形式启动 & 才能起作用。
#直接在终端中敲下面的命令关掉终端，进程还是结束了。。

#捕获崩溃异常，dlv
ulimit -c unlimited
export GOTRACEBACK=crash

#多个服务只是目录的不同
#必须重定向到/dev/null，否则远程启动时，关闭终端，进程也会结束
$(pwd)/exeNameReplace &
#$(pwd)/exeNameReplace > /dev/null &
"""

    def get_end() -> str:
        return """#!/bin/bash

#通用停止当前目录启动的进程
#kill -9 发送 SIGKILL信号
#ps -aux | grep "$(pwd)" | grep -v "grep" | awk '{print $2}' | xargs kill -9
#15 发送 SIGTERM信号，允许程序优雅退出
#ps -aux | grep "$(pwd)" | grep -v "grep" | awk '{print $2}' | xargs kill -15

function stop(){
    pid=$(ps -aux | grep "$(pwd)/$1" | grep -v "grep" | awk '{print $2}')
    if [ -n "$pid" ]; then
        echo "stop $1 pid: $pid"
        echo "$pid" | sudo xargs kill "$2"
    else
        echo "No process:$1 to kill"
    fi
}

function stop15(){
    stop "$1" -15
}

stop15 exeNameReplace
"""

    def get_restart() -> str:
        return """#!/bin/bash

#上传后，如果运行报错。更改回车符：goland->File->File Properties->Line Separators->LF-Unix & macOs

#通用重启当前目录的进程。
sudo chmod +xxx $(pwd)/start.sh
sudo chmod +xxx $(pwd)/end.sh

sudo $(pwd)/end.sh
echo "sleep 2s..."
sleep 2s
sudo $(pwd)/start.sh
"""

    def get_screen() -> str:
        return """#退出对应的screen  ||true 忽略执行错误
screen -S screenNameReplace -X quit || true
#重新创建一个新的screen
screen -dmS screenNameReplace || false

#指定执行脚本内容
script="cd linuxDirReplace  && ./start.sh"
#离屏执行一段内容
screen -S screenNameReplace -X eval "screen" "-X" "stuff '${script} \n'"
"""


class Setup(BaseSetup):
    def __init__(self, args: object | None = None) -> None:
        super().__init__(
            args and args.ip,
            args and args.port,
            args and args.user,
            args and args.vpn_host,
            args and args.vpn_port,
        )

        self.dir = args and args.dir or "/opt/test"
        self.exe = args and args.exe or "exe"
        self.screen = args and args.screen or ""
        self.outDir = args and args.out_dir or "."
        self.upload_bin = args and args.upload_bin
        self.buildvcs = args and args.buildvcs or False

        if args:
            self.buildDir = args.build_dir
            self.buildFile = args.build_file

            self.shellOn = args.shell
            self.sudo = args.sudo and "sudo " or ""
            self.shellStart = args.shell_start

            self.cacheDirNames = args.cache_dir_names
            self.envs = args.envs
            self.crontab = args.crontab

    def grant_privilege(self, targetDir: str) -> bool:
        # 给目录权限
        if self.user != "root":
            if not self.remote_exec(f"sudo chown {self.user} {targetDir}"):
                return False
        if not self.remote_exec(f"sudo chmod 751 {targetDir}"):
            return False
        return True

    def upload_pre(self) -> bool:
        # 编译代码
        if self.upload_bin:
            if not build_go(
                self.buildDir,
                self.buildFile,
                self.outDir,
                self.exe,
                self.buildvcs,
            ):
                return False

        # 创建服务器目录环境
        if not self.remote_exec(f"sudo mkdir -p {self.dir}"):
            return False

        if not self.grant_privilege(self.dir):
            return False

        return True

    def upload_after(self) -> bool:
        # 最后上传EXE
        exeFullPath = self.outDir + "/" + self.exe
        my_print(f"上传EXE,觉得慢可以直接终止! ip:{self.ip} exe:{exeFullPath}")
        zipName = self.exe + ".zip"
        zipFullPath = self.outDir + "/" + zipName

        # 完全没有压缩。。。
        # # 压缩exe文件
        # with ZipFile(zipFullPath, "w") as myzip:
        #     myzip.write(exeFullPath, self.exe)

        # # 上传exe
        # if not self.remote_put(zipFullPath, self.dir):  # + "/" + self.exe
        #     return False

        # # 解压exe
        # if not self.remote_exec(f"cd {self.dir} && unzip -o {zipName}"):
        #     return False

        if not self.remote_put(exeFullPath, self.dir):  # + "/" + self.exe
            return False
        return True

    def upload(self) -> bool:
        if not self.upload_pre():
            return False
        return self.upload_after()

    def copy_pre(self) -> bool:
        for i in range(len(self.cacheDirNames)):
            subDir = self.dir + "/" + self.cacheDirNames[i]
            if not self.remote_exec(f"sudo mkdir -p {subDir}"):
                return False
            if not self.grant_privilege(subDir):
                return False

        # # 重命名
        if self.upload_bin:
            # self.remote_exec(f"sudo cd {self.dir} && rm {self.exe}.bak")
            self.remote_exec(f"sudo mkdir -p {self.dir}/back")
            # self.remote_exec(f'sudo find {self.dir}/back -mtime +30 -type f -name "*.bak" -delete')
            self.remote_exec(
                f"sudo mv {self.dir}/{self.exe} {self.dir}/back/{self.exe}.{int(time.time())}.bak"
            )

        # 上传文件
        if self.shellOn:
            # 上传
            shells = [
                {"name": "end.sh", "content": BaseSetup.get_end()},
                {"name": "start.sh", "content": BaseSetup.get_start()},
                {"name": "restart.sh", "content": BaseSetup.get_restart()},
                {"name": "screen.sh", "content": BaseSetup.get_screen()},
            ]
            for i in range(len(shells)):
                thePath = f'{self.outDir}/{shells[i]["name"]}'
                with open(thePath, "wb") as f:
                    f.write(shells[i]["content"].encode("utf8"))
                if not self.remote_put(thePath, self.dir):  # + "/" + shells[i]
                    return False
                os.remove(thePath)

        if self.shellStart != "":
            if not self.remote_put(self.shellStart, self.dir):  # + "/start.sh"
                return False

        for i in range(0, len(self.envs), 3):
            here, there, isDir = self.envs[i], self.envs[i + 1], self.envs[i + 2] == "1"
            if not self.remote_put(here, self.dir + "/" + there, isDir):
                return False

        # 调整替换顺序，支持 envs 配置自定义start.sh脚本文件，并支持模板更改其中的名字
        if self.shellOn:
            # 替换
            if not self.remote_exec(
                f'sudo sed -i "s/exeNameReplace/{self.exe}/g" {self.dir}/start.sh'
            ):
                return False

            if not self.remote_exec(
                f'sudo sed -i "s/exeNameReplace/{self.exe}/g" {self.dir}/end.sh'
            ):
                return False

            screenName = self.exe + self.screen
            if not self.remote_exec(
                f'sudo sed -i "s/screenNameReplace/{screenName}/g" {self.dir}/screen.sh'
            ):
                return False
            if not self.remote_exec(
                f'sudo sed -i "s#linuxDirReplace#{self.dir}#g" {self.dir}/screen.sh'
            ):
                return False
        return True

    def copy_after(self) -> bool:
        if self.shellOn:
            # 关闭
            if not self.remote_exec(f"cd {self.dir} && sudo chmod +xxx ./restart.sh"):
                return False
            if not self.remote_exec(
                f"cd {self.dir} && sudo chmod +xxx ./end.sh && ./end.sh"
            ):
                return False

            # sleep 2秒
            my_print("sleep 2...")
            time.sleep(2)

            # 启动
            if not self.remote_exec(f"cd {self.dir} && sudo chmod +xxx ./start.sh"):
                return False
            if not self.remote_exec(
                f"cd {self.dir} && sudo chmod +xxx ./screen.sh && ./screen.sh"
            ):
                return False

        # 定时任务配置
        if self.crontab != "":
            crontabPath = self.dir + "/" + self.exe + ".crontab"
            if not self.remote_put(self.crontab, crontabPath):
                return False
            if not self.remote_exec(f"sudo crontab {crontabPath}"):
                return False

        my_print(f"setup finished {self.ip}:{self.dir}/{self.exe}")
        return True

    def copy_start(self, srcBin: str) -> bool:
        if not self.copy_pre():
            return False

        # copy
        if not self.remote_exec(f"sudo cp {srcBin} {self.dir}"):
            return False

        return self.copy_after()

    def start(self) -> bool:
        if not self.upload_pre():
            return False

        if not self.copy_pre():
            return False

        if self.upload_bin:
            # 最后上传EXE
            if not self.upload_after():
                return False

        return self.copy_after()


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
    parser.add_argument("-build_dir", help="编译目录")  # 与编译文件二选一
    parser.add_argument("-build_file", help="编译文件")  # 与编译目录二选一
    parser.add_argument("-out_dir", required=True, help="编译输出目录")
    parser.add_argument("-user", default="root", help="远程用户")

    # 这个True False的只能用两个减
    parser.add_argument(
        "--shell", action=argparse.BooleanOptionalAction, help="是否上传通用shell脚本"
    )
    parser.add_argument(
        "--sudo", action=argparse.BooleanOptionalAction, help="执行命令前sudo吗"
    )
    parser.add_argument(
        "-cache_dir_names", nargs="*", help="远程备用目录名,以dir指定的目录为当前前缀"
    )
    parser.add_argument(
        "-envs", nargs="*", help="环境配置文件。例:本地文件名,服务器文件名,1目录/0文件"
    )
    parser.add_argument("-crontab", default="", help="远程用户")
    parser.add_argument("-shell_start", default="", help="特殊的启动命令shell")
    parser.add_argument("-vpn_host", default="localhost", help="代理地址")
    parser.add_argument("-vpn_port", default=0, type=int, help="代理端口")
    parser.add_argument("-action", default="start", help="执行行为")
    parser.add_argument("-src", default="", help="拷贝执行的原二进制远程路径")
    parser.add_argument(
        "--upload_bin",
        action=argparse.BooleanOptionalAction,
        default=True,
        help="是否上传二进制",
    )
    parser.add_argument(
        "--buildvcs",
        action=argparse.BooleanOptionalAction,
        default=False,
        help="是否buildvcs",
    )

    args = parser.parse_args()
    my_print("parseArgs=", args)
    setup = Setup(args)

    print("action=", args.action)
    if args.action == "start":
        if not setup.start():
            raise RuntimeError("exec fail")
    elif args.action == "upload":
        if not setup.upload():
            raise RuntimeError("exec fail")
    elif args.action == "copy_start":
        if args.src == "":
            raise ValueError("copy src is empty")
        if not setup.copy_start(args.src):
            raise RuntimeError("exec fail")
    else:
        raise ValueError("unknow action " + args.action)


def call_setup(kvargs: dict, envs: list, py: str = "setup.py"):
    args = ["python", py]
    for k, v in kvargs.items():
        args.append(k)
        if type(v) == list:
            for i in range(len(v)):
                args.append(str(v[i]))
        elif v != "":
            args.append(str(v))

    if envs:
        args.append("-envs")
        for i in range(len(envs)):
            args.append(envs[i]["from"])
            args.append(envs[i]["to"])
            if "dir" in envs[i] and envs[i]["dir"]:
                args.append("1")
            else:
                args.append("0")

    print("call_setup args=", args)

    try:
        subprocess.check_call(args)
    except Exception as e:
        print("call_setup e=", e)
        # 终止，异常
        os._exit(1)


def test():
    # remote_exec("127.0.0.1", "ls -l")
    # remote_exec("127.0.0.1", "pwd")
    # cwd = os.getcwd()  # 获取当前目录
    s = BaseSetup.get_acme("www.baidu.com")
    print(s)
    remote_exec("127.0.0.1", f"echo '{s}' > /root/acme.test.sh")
    # remote_put("127.0.0.1", cwd + "/../auth/cfg/yqw2.json", "/opt/auth/auth.json")

    pass


if __name__ == "__main__":
    # test()
    # print(__file__, os.path.dirname(__file__)
    parse_to_setup()

    # print(BaseSetup.get_acme("1.cn"))
