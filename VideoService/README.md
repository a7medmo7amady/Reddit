# Video Service (Express.js + Kafka)

A scalable video processing service that handles uploads, transcoding (HLS), and metadata management.

## Features
- **REST API**: Create video records, complete uploads, and fetch metadata.
- **Event-Driven**: Uses Kafka for decoupling transcoding from the upload flow.
- **FFmpeg Integration**: Automatic thumbnail generation and HLS packaging.
- **S3 Integration**: Processed assets are stored in S3/MinIO.

## Prerequisites
- Node.js (v16+)
- Kafka (running on localhost:9092)
- MinIO (running on localhost:9000)
- FFmpeg installed in the system PATH

## Installation
```bash
cd VideoService
npm install
```

## Running the Service
```bash
npm start
```

## API Endpoints
- `POST /videos`: Create a video record.
- `POST /videos/:id/complete`: Finalize upload and trigger transcoding.
- `GET /videos/:id`: Get video metadata and playback URLs.
- `GET /videos/:id/status`: Get current processing status.

## Architecture
1. Client calls `POST /videos` to get a `videoId`.
2. Client uploads file to S3.
3. Client calls `POST /videos/:id/complete`.
4. Service publishes `video.uploaded` event to Kafka.
5. Transcoding worker consumes the event, runs FFmpeg, and uploads HLS segments.
6. Service publishes `video.ready` upon completion.
