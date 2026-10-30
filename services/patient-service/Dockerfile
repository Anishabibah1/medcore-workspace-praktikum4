# STAGE 1: Build & Transpilation
FROM node:20-alpine3.19 AS builder
WORKDIR /usr/src/app
COPY package*.json tsconfig.json ./
RUN npm ci
COPY src/ ./src/
RUN npm run build && npm prune --omit=dev

# STAGE 2: Production Runner
FROM node:20-alpine3.19 AS runner
WORKDIR /app
ENV NODE_ENV=production
USER 1000:1000
COPY --chown=node:node package*.json ./
COPY --chown=node:node --from=builder /usr/src/app/node_modules ./node_modules
COPY --chown=node:node --from=builder /usr/src/app/dist ./dist
EXPOSE 3000
CMD ["node", "dist/index.js"]
