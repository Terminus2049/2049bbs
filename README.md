2049bbs，一个无需手机号和邮箱即可注册发言的论坛。Fork 自 [goyoubbs](https://github.com/ego008/goyoubbs)。

## 本地开发


安装 [go](https://golang.org/dl/)，然后 clone 本仓库。

```bash
go get -v github.com/terminus2049/2049bbs
```
1. 然后 cd 到相应目录，一般是 `go/src/github.com/terminus2049/2049bbs`。

```bash
go run main.go
```

然后在浏览器打开 `127.0.0.1:8082` 即可，或者直接编译，运行 `sudo ./2049bbs`，

2. 利用 Docker 进行开发

 - 首先安装 Docker 及 docker-compose
 - 将本项目 clone 到本地，任何目录均可
 - 进入项目目录，在安装好 docker 及 docker-compose 后，运行脚本 `make dev` 即自动拉去构建好的镜像
 - 运行成功后 `docker ps` 即可发现名为 bbs 的容器正在运行中

```
machine: 2049BBS % docker ps
CONTAINER ID        IMAGE                                                                       COMMAND                  CREATED             STATUS              PORTS                    NAMES
d798030a6f0f        docker.pkg.github.com/speechfree/go-base/go-base:base                       "tail -f /dev/null"      About an hour ago   Up About an hour    0.0.0.0:8000->8082/tcp   bbs
```
 - 然后，`docker exec -it bbs bash` 进入到容器中，通过 `dep ensure` 拉去项目依赖到本地目录 `vendor` 中
 - 完成后运行 `go run main.go` 若出现如下输出即表明项目运行成功

```
2019/12/20 13:23:07 MainDomain: http://127.0.0.1:8082
2019/12/20 13:23:07 youdb Connect to mydata.db
2019/12/20 13:23:07 Web server Listen port 8082
```
 - 在宿主机打开任意浏览器输入 `http://localhost:8000` 即可看到构建成功的应用。
 - 另，在开发过程中，为了方便修改代码后重载应用，可以通过 `realize start` 启动应用，则修改任何 Golang 代码，其均会自动构建加载。

### 数据库

如果没有 kv 数据库开发经验，最好在程序跑起来后，用 [boltdbweb](https://github.com/evnix/boltdbweb) 打开数据库文件 `mydata.db`，了解一下内部存储结构。

## 部署

编译二进制文件 `go build`，~~非 Linux 平台为交叉编译 `GOOS=linux GOARCH=amd64 go build`~~，由于使用了 gojieba 分词引擎，不能跨平台编译，请使用在线api功能、移除相关组件后再尝试跨平台编译。

将编译好的二进制文件与 config、static 和 view 三个文件夹的文件放在同一个文件夹内，运行 `./2049bbs`。

服务器配置：在生产环境中，建议打开 `https`，把 `config.yaml` 中 `HttpsOn: false` 改为 `true`。也可以自行申请 cloudflare 证书，相应配置可以参考 [config-2049.yaml](https://github.com/Terminus2049/2049BBS/blob/master/config/config-2049.yaml).

## 备份

需要备份 `mydata.db` 和 `/static/avatar` 文件。
