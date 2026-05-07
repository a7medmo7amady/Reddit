require('dotenv').config();
const express = require('express');
const mongoose = require('mongoose');
const videoRoutes = require('./routes/video.js');
const kafkaService = require('./services/kafka.service');
const transcoderService = require('./services/transcoder.service');
const storageService = require('./services/storage.service');
const VideoModel = require('./models/video.model');


const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());
app.use('/', videoRoutes);
app.get('/health', (req, res) => res.json({ status: 'OK' }));

async function start() {
    app.listen(PORT, '0.0.0.0');

    try {
        await mongoose.connect(process.env.MONGODB_URI || 'mongodb://localhost:27017/video_service');
        await storageService.initialize();
        await kafkaService.connect();
        await kafkaService.createTopics(['video.uploaded', 'video.processing', 'video.ready']);

        await kafkaService.subscribe('video.uploaded', async (payload) => {
            const { videoId, s3Key } = payload;
            await VideoModel.update(videoId, { status: 'PROCESSING' });
            await kafkaService.publish('video.processing', { videoId });
            await transcoderService.processVideo(videoId, s3Key);
        });

        await kafkaService.subscribe('video.ready', async (payload) => {});
        await kafkaService.startConsumer();
    } catch (error) {
        console.error(error.message);
    }
}

start();

const shutdown = async () => {
    await kafkaService.disconnect();
    process.exit(0);
};
process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);
