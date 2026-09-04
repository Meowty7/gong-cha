FROM oven/bun:1.4.0-alpine AS builder
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
# Empty = same-origin `/api` (nginx or Caddy proxy). Override only if the API is on another host.
ARG PUBLIC_API_URL=
ENV PUBLIC_API_URL=$PUBLIC_API_URL
RUN bun run build

FROM nginx:1.28-alpine
RUN apk add --no-cache wget
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
