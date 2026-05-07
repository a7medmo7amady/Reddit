const ffmpeg = require('fluent-ffmpeg');
const path = require('path');
const fs = require('fs');
const storageService = require('./storage.service');
const kafkaService = require('./kafka.service');
const VideoModel = require('../models/video.model');

class TranscoderService {
    async processVideo(videoId, rawS3Key) {
        const video = await VideoModel.findById(videoId);
        if (video && video.status === 'READY') return;

        const tempDir = path.join(__dirname, '../temp', videoId);
        if (!fs.existsSync(tempDir)) fs.mkdirSync(tempDir, { recursive: true });
        const rawLocalPath = path.join(tempDir, 'raw.mp4');

        try {
            await storageService.downloadFile(rawS3Key, rawLocalPath, 'staging');

            const thumbPath = await this.generateThumbnail(rawLocalPath, tempDir);
            await storageService.uploadFile(thumbPath, `thumbs/${videoId}.jpg`, 'serving');

            const gifPath = await this.generatePreviewGif(rawLocalPath, tempDir);
            await storageService.uploadFile(gifPath, `previews/${videoId}.gif`, 'serving');

            const resolutions = await this.transcodeToHLS(rawLocalPath, tempDir);

            const files = fs.readdirSync(tempDir);
            for (const file of files) {
                if (file.endsWith('.m3u8') || file.endsWith('.ts')) {
                    await storageService.uploadFile(path.join(tempDir, file), `video/${videoId}/${file}`, 'serving');
                }
            }

            await VideoModel.update(videoId, { 
                status: 'READY', 
                thumbnailUrl: `/thumbs/${videoId}.jpg`,
                previewUrl: `/previews/${videoId}.gif`,
                playbackUrl: `/video/${videoId}/master.m3u8`,
                resolutions: resolutions
            });
            await kafkaService.publish('video.ready', { videoId, status: 'READY' });
        } catch (error) {
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
        const configs = [
            { resolution: '360p', size: '640x360', bitrate: '800k' },
            { resolution: '720p', size: '1280x720', bitrate: '2500k' }
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

        // Create Master Playlist
        const masterPlaylistContent = [
            '#EXTM3U',
            '#EXT-X-VERSION:3',
            '#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360',
            '360p.m3u8',
            '#EXT-X-STREAM-INF:BANDWIDTH=2500000,RESOLUTION=1280x720',
            '720p.m3u8'
        ].join('\n');

        fs.writeFileSync(path.join(outputDir, 'master.m3u8'), masterPlaylistContent);

        return configs.map(c => c.resolution);
    }
}

module.exports = new TranscoderService();
