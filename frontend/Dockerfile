# Build image
FROM node:18-alpine AS build
WORKDIR /app
COPY package.json vite.config.js index.html ./
COPY src ./src
RUN npm install
RUN npm run build

# Production image
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
