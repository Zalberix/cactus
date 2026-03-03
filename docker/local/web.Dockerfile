FROM node:24-alpine

WORKDIR /app

# Install dependencies as a separate layer for caching
COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
RUN npm install

EXPOSE 3000

# Required for Next.js dev server to bind to all interfaces inside container
ENV HOSTNAME="0.0.0.0"

CMD ["npm", "run", "dev"]
