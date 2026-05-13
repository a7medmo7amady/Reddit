const ffmpeg = require('fluent-ffmpeg');
const path = require('path');
const fs = require('fs');
const storageService = require('./storage.service');
const kafkaService = require('./kafka.service');
const PostModel = require('../models/post.model');

class TranscoderService {
    async processVideo(postId, rawS3Key) {
        const post = await PostModel.findById(postId);
        if (post && post.video?.status === 'READY') return;

        const tempDir = path.join(__dirname, '../temp', postId);
        if (!fs.existsSync(tempDir)) fs.mkdirSync(tempDir, { recursive: true });
        const rawLocalPath = path.join(tempDir, 'raw.mp4');

        try {
            await storageService.downloadFile(rawS3Key, rawLocalPath, 'staging');

            // ── Thumbnail ──────────────────────────────────────────────────
            const thumbPath = await this.generateThumbnail(rawLocalPath, tempDir);
            await storageService.uploadFile(thumbPath, `thumbs/${postId}.jpg`, 'serving');

            // ── Preview GIF ────────────────────────────────────────────────
            const gifPath = await this.generatePreviewGif(rawLocalPath, tempDir);
            await storageService.uploadFile(gifPath, `previews/${postId}.gif`, 'serving');

            // ── HLS Transcoding (360p, 720p, 1080p) ───────────────────────
            const resolutions = await this.transcodeToHLS(rawLocalPath, tempDir);

            // ── Upload all HLS segments & manifests ────────────────────────
            const files = fs.readdirSync(tempDir);
            for (const file of files) {
                if (file.endsWith('.m3u8') || file.endsWith('.ts')) {
                    await storageService.uploadFile(
                        path.join(tempDir, file),
                        `video/${postId}/${file}`,
                        'serving'
                    );
                }
            }

            await PostModel.update(postId, {
                'video.status':       'READY',
                'video.thumbnailUrl': `/assets/thumbs/${postId}.jpg`,
                'video.previewUrl':   `/assets/previews/${postId}.gif`,
                'video.playbackUrl':  `/assets/video/${postId}/master.m3u8`,
                'video.resolutions':  resolutions,
            });

            await kafkaService.publish('video.ready', { postId, status: 'READY' });
        } catch (error) {
            await PostModel.update(postId, { 'video.status': 'FAILED' });
            throw error;
        } finally {
            if (fs.existsSync(tempDir)) fs.rmSync(tempDir, { recursive: true, force: true });
        }
    }

    generateThumbnail(inputPath, outputDir) {
        return new Promise((resolve, reject) => {
            ffmpeg(inputPath)
                .screenshots({
                    timestamps: [3],
                    filename: 'thumb.jpg',
                    folder: outputDir,
                    size: '640x360'
                })
                .on('end', () => resolve(path.join(outputDir, 'thumb.jpg')))
                .on('error', reject);
        });
    }

    generatePreviewGif(inputPath, outputDir) {
        const outputPath = path.join(outputDir, 'preview.gif');
        return new Promise((resolve, reject) => {
            ffmpeg(inputPath)
                .setStartTime(3)
                .setDuration(3)
                .fps(10)
                .size('320x180')
                .output(outputPath)
                .on('end', () => resolve(outputPath))
                .on('error', reject)
                .run();
        });
    }

    async transcodeToHLS(inputPath, outputDir) {
        // FR-UC-03: 360p, 720p, 1080p variants
        const configs = [
            { resolution: '360p',  size: '640x360',   bitrate: '800k'  },
            { resolution: '720p',  size: '1280x720',  bitrate: '2500k' },
            { resolution: '1080p', size: '1920x1080', bitrate: '5000k' },
        ];

        for (const config of configs) {
            await new Promise((resolve, reject) => {
                ffmpeg(inputPath)
                    .size(config.size)
                    .videoBitrate(config.bitrate)
                    .outputOptions([
                        '-profile:v baseline',
                        '-level 3.0',
                        '-start_number 0',
                        '-hls_time 10',
                        '-hls_list_size 0',
                        '-f hls'
                    ])
                    .output(path.join(outputDir, `${config.resolution}.m3u8`))
                    .on('end', resolve)
                    .on('error', reject)
                    .run();
            });
        }

        // Master playlist referencing all three variants
        const masterPlaylistContent = [
            '#EXTM3U',
            '#EXT-X-VERSION:3',
            '#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360',
            '360p.m3u8',
            '#EXT-X-STREAM-INF:BANDWIDTH=2500000,RESOLUTION=1280x720',
            '720p.m3u8',
            '#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080',
            '1080p.m3u8',
        ].join('\n');

        fs.writeFileSync(path.join(outputDir, 'master.m3u8'), masterPlaylistContent);

        return configs.map(c => c.resolution);
    }
}

module.exports = new TranscoderService();
