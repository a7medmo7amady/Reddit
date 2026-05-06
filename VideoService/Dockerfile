FROM node:20-alpine

# Install FFmpeg
RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

# Create temp directories for transcoding and uploads
RUN mkdir -p temp/uploads && chmod -R 777 temp

EXPOSE 3000

CMD ["npm", "start"]
