# 使用 golang 1.20.3 作为基础镜像官方镜像
FROM golang:1.20.3 as builder

# 将工作目录切换到程序代码所在的目录
WORKDIR /app

# 将当前目录下的所有文件都复制到工作目录下
COPY . .

# 设置proxy
RUN go env -w GOPROXY=https://goproxy.cn,direct

# 构建Go程序
RUN go build -o /app/pkg/gpu/job-server /app/pkg/gpu/main

# 拷贝job-server到根目录
RUN cp /app/pkg/gpu/job-server /bin/job-server

# 基础镜像ubuntu
FROM ubuntu:20.04

# 将构建的job-server文件复制到Ubuntu镜像中
COPY --from=builder /app/pkg/gpu/job-server /bin/job-server


# 启动Go程序
ENTRYPOINT ["/bin/job-server"]


# 构建镜像
# 要构建容器，可以使用以下命令：
# 需要在项目的根路径执行
# docker build -t job-server:latest .
# docker run --entrypoint /bin/job-server musicminion/minik8s-gpu -jobName job-example1 -jobNamespace test-job-namespace -apiServerAddr http://192.168.126.130:8090