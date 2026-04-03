# Stage 1: build iframe bundle (dist/assets is not part of ksbuilder create output).
FROM node:20-alpine AS assets
WORKDIR /build
COPY package.json package-lock.json ./
RUN npm ci
COPY src ./src
RUN npm run build:assets && npm run build:shell

FROM nginx:alpine
COPY --from=assets /build/dist/assets/bundle.js /usr/share/nginx/html/assets/bundle.js
COPY index.html /usr/share/nginx/html/index.html
COPY --from=assets /build/dist/index.js /usr/share/nginx/html/dist/index.js
EXPOSE 80
