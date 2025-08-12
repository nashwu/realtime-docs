# Build stage 
FROM node:20-alpine AS build
WORKDIR /app

# Install deps with maximum cache reuse
COPY package*.json ./
RUN npm ci

COPY . .
RUN npx ng build --configuration production

# Runtime stage
FROM nginx:1.27-alpine
WORKDIR /usr/share/nginx/html

# Replace default nginx site with SPA + API/WS proxy
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/app.conf

COPY --from=build /app/dist/frontend/browser/ /usr/share/nginx/html/

EXPOSE 80
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1/ >/dev/null 2>&1 || exit 1

CMD ["nginx", "-g", "daemon off;"]
