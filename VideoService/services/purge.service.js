const cron = require('node-cron');
const PostModel = require('../models/post.model');
const storageService = require('./storage.service');

class PurgeService {
    start() {
        // Runs every hour at minute 0
        cron.schedule('0 * * * *', async () => {
            await this._purge();
        });
        console.log('[PurgeService] Scheduled: deleted media purged every hour after 24h.');
    }

    async _purge() {
        try {
            const posts = await PostModel.findPurgeable();
            if (posts.length === 0) return;

            console.log(`[PurgeService] Purging ${posts.length} deleted post(s)...`);

            for (const post of posts) {
                await this._purgePost(post);
            }
        } catch (err) {
            console.error('[PurgeService] Error during purge run:', err.message);
        }
    }

    async _purgePost(post) {
        const servingBucket = process.env.S3_SERVING_BUCKET || 'serving';
        const imagesBucket  = process.env.S3_IMAGES_BUCKET  || 'images';
        const stagingBucket = process.env.S3_STAGING_BUCKET || 'staging';

        try {
            if (post.type === 'video' && post.video) {
                // Delete raw staging file
                if (post.video.s3Key) {
                    await storageService.deleteFile(post.video.s3Key, stagingBucket).catch(() => {});
                }

                // Delete HLS segments & manifests stored as video/<id>/*
                const hlsPrefix = `video/${post.id}/`;
                await storageService.deleteFolder(hlsPrefix, servingBucket).catch(() => {});

                // Delete thumbnail and preview GIF
                await storageService.deleteFile(`thumbs/${post.id}.jpg`, servingBucket).catch(() => {});
                await storageService.deleteFile(`previews/${post.id}.gif`, servingBucket).catch(() => {});
            }

            if (post.type === 'image' && post.images?.length > 0) {
                for (const img of post.images) {
                    const variants = [img.thumbnail, img.preview, img.full].filter(Boolean);
                    for (const url of variants) {
                        // URL pattern: /assets/images/<postId>/<filename>
                        const s3Key = url.replace(/^\/assets\/images\//, 'images/');
                        await storageService.deleteFile(s3Key, imagesBucket).catch(() => {});
                    }
                }
            }

            // Mark as purged so this post is never re-processed by the cron
            await PostModel.update(post.id, { purged: true });
            console.log(`[PurgeService] Purged post ${post.id} (type: ${post.type})`);
        } catch (err) {
            console.error(`[PurgeService] Failed to purge post ${post.id}:`, err.message);
        }
    }
}

module.exports = new PurgeService();
