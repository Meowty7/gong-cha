FROM oven/bun:1.4.0-alpine AS builder
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
ARG PUBLIC_API_URL=https://gong-cha.belta.dev
ENV PUBLIC_API_URL=$PUBLIC_API_URL
RUN bun run build

FROM nginx:1.28-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
