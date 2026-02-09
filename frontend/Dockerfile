# 1. Build Stage
FROM node:18-alpine as builder
WORKDIR /app

# Cache dependencies
COPY package.json package-lock.json ./
RUN npm install

# Copy source code (this layer will invalidate when code changes)
COPY . .

# Force rebuild by adding a build arg
ARG CACHEBUST=1

# Build the app
RUN npm run build

# 2. Run Stage
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]