# 使用 golang 1.20.3 作为基础镜像官方镜像
FROM golang:1.20.3

# 将工作目录切换到程序代码所在的目录
WORKDIR /app

# 将当前目录下的所有文件都复制到工作目录下
COPY . .


# 设置proxy
RUN go env -w GOPROXY=https://goproxy.cn,direct

# 构建Go程序
RUN go build -o /app/pkg/gpu/job-server /app/pkg/gpu/main

# 拷贝job-server到根目录
RUN cp ./job-server /bin/job-server

# 暴露端口如果需要
# EXPOSE 8080

# 启动Go程序
ENTRYPOINT ["/bin/job-server"]

# 构建镜像
# 要构建容器，可以使用以下命令：
# docker build -t job-server:latest .
