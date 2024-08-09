# syntax=docker/dockerfile:1

FROM node:22 AS builder
ENV NODE_ENV=production
WORKDIR /frontend

COPY ./frontend/package.json ./frontend/package-lock.json /frontend/
RUN npm install
COPY ./frontend /frontend
RUN npm run generate

FROM nginx:1.27.0-alpine AS runner
RUN apk --update add tzdata

# --from=builder dist -> /frontend/.output/public
COPY --from=builder /frontend/.output/public /usr/share/nginx/html/
RUN rm -f /usr/share/nginx/html/200.html
