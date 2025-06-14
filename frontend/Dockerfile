# 构建阶段
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 配置国内镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置 npm 国内镜像源
RUN npm config set registry https://registry.npmmirror.com/

# 复制 package 文件
COPY package*.json ./

# 清理npm缓存并安装所有依赖
RUN npm cache clean --force && \
    rm -rf node_modules package-lock.json && \
    npm install

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# 运行阶段
FROM nginx:alpine

# 配置国内镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 复制构建好的文件到 nginx 目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 80

# 启动 nginx
CMD ["nginx", "-g", "daemon off;"] 