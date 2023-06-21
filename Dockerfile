# 基础镜像
FROM golang:1.19

# 设置工作目录
WORKDIR /app

# 将本地文件复制到容器中
COPY . .

# 构建应用
RUN go build -o main .

# 设置环境变量
ENV PORT=8000

# 暴露端口
EXPOSE $PORT

# 运行应用
CMD ["./main"]