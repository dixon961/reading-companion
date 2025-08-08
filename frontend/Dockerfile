# Use node:alpine as base image
FROM node:20.19.0-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install
RUN npm ci

# Copy source code
COPY . .

# Expose port 5173 (default for Vite dev server)
EXPOSE 5173

# Command to run the application in DEV!
CMD ["npm", "run", "dev"]