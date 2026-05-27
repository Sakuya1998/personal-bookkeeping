FROM mcr.microsoft.com/playwright:v1.60.0-noble

WORKDIR /e2e

# 先装依赖（利用 Docker layer 缓存）
COPY package.json package-lock.json ./
RUN npm ci

# 复制测试代码
COPY . .

# 默认入口：API 测试
CMD ["npx", "playwright", "test", "--project=api"]
