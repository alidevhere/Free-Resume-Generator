FROM node:20-alpine

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

ENV BACKEND_API_URL=http://backend:8080

RUN npm run build

EXPOSE 3000

CMD ["npm", "run", "start"]
