# Frontend project setup
FROM node:20-alpine

WORKDIR /app

# Set build-time environment variables
ENV NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_WS_URL=ws://localhost:8080
ENV NEXT_TELEMETRY_DISABLED=1

COPY package*.json ./

RUN npm install --legacy-peer-deps

COPY . .

# Build with the API URL baked in
RUN npm run build

EXPOSE 3000

CMD ["npm", "start"]
