# 使用 Node.js 官方的 Alpine 镜像作为基础镜像
FROM node:alpine as build

# 设置工作目录
WORKDIR /app

# 将项目文件复制到工作目录
COPY . /app

# 安装项目依赖
RUN yarn install

# 构建生产版本
RUN yarn build

# 使用 Nginx 官方的 Alpine 镜像作为基础镜像
FROM nginx:alpine

# 复制构建好的静态文件到 Nginx 的默认目录
COPY --from=build /app/build /usr/share/nginx/html

# 暴露 Nginx 默认端口
EXPOSE 80

# 容器启动时运行 Nginx
CMD ["nginx", "-g", "daemon off;"]
