require('dotenv').config();
const express = require('express');
const mongoose = require('mongoose');

const postRoutes    = require('./routes/posts');
const commentRoutes = require('./routes/comments');
const assetRoutes   = require('./routes/assets');
const kafkaService  = require('./services/kafka.service');
const transcoderService = require('./services/transcoder.service');
const storageService    = require('./services/storage.service');
const purgeService  = require('./services/purge.service');
const seedService   = require('./services/seed.service');
const PostModel     = require('./models/post.model');
const consulService = require('./services/consul.service');

const app  = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());
app.use('/', postRoutes);
app.use('/', commentRoutes);
app.use('/', assetRoutes);
app.get('/health', (req, res) => res.json({ status: 'OK', service: 'video-service' }));

async function start() {
    app.listen(PORT, '0.0.0.0');
    console.log(`[VideoService] Listening on port ${PORT}`);

    try {
        await mongoose.connect(process.env.MONGODB_URI || 'mongodb://localhost:27017/upload_service');
        console.log('[VideoService] MongoDB connected.');

        await storageService.initialize();
        console.log('[VideoService] Storage buckets ready (staging, serving, images).');

        await kafkaService.connect();
        await kafkaService.createTopics(['video.uploaded', 'video.processing', 'video.ready', 'post']);

        // Video uploaded → begin transcoding pipeline
        await kafkaService.subscribe('video.uploaded', async (payload) => {
            const { postId, s3Key } = payload;
            await PostModel.update(postId, { 'video.status': 'PROCESSING' });
            await kafkaService.publish('video.processing', { postId });
            // Run in background so Kafka consumer isn't blocked and heartbeats continue
            transcoderService.processVideo(postId, s3Key).catch(err =>
                console.error(`[Transcoder] postId=${postId} failed:`, err.message)
            );
        });

        await kafkaService.subscribe('video.ready', async () => {});
        await kafkaService.startConsumer();
        console.log('[VideoService] Kafka consumer started.');

        await seedService.seed();

        // Start 24h media purge cron
        purgeService.start();

        // Register with Consul for service discovery
        await consulService.register();

    } catch (error) {
        console.error('[VideoService] Startup error:', error.message);
    }
}

start();

const shutdown = async () => {
    console.log('[VideoService] Shutting down...');
    await consulService.deregister();
    await kafkaService.disconnect();
    process.exit(0);
};

// Global error handler for middleware (like Multer)
app.use((err, req, res, next) => {
    console.error('[GlobalError]', err.stack || err.message);
    res.status(500).json({ error: 'Internal Server Error', details: err.message });
});

process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);
