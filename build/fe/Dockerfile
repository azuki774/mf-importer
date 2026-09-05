# syntax=docker/dockerfile:1

FROM node:24-alpine AS builder
WORKDIR /frontend

COPY ./frontend/package.json ./frontend/package-lock.json /frontend/
RUN npm ci
COPY ./frontend /frontend
RUN npm run build

FROM node:24-alpine AS runner
RUN apk --update add tzdata

COPY --from=builder /frontend/.output/ ./src/.output
COPY --from=builder /frontend/node_modules ./src/node_modules
COPY --from=builder /frontend/package.json ./src/package.json

EXPOSE 3000
CMD ["node", "src/.output/server/index.mjs"]
