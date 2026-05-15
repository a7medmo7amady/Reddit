const express = require('express');
const router = express.Router();
const { v4: uuidv4 } = require('uuid');
const multer = require('multer');
const fs = require('fs');
const mongoose = require('mongoose');

const PostModel = require('../models/post.model');
const kafkaService = require('../services/kafka.service');
const storageService = require('../services/storage.service');
const imageService = require('../services/image.service');
const banService = require('../services/ban.service');
const consulService = require('../services/consul.service');
const { Upload } = require('@aws-sdk/lib-storage');

// ── Multer config ─────────────────────────────────────────────────────────────
// Single disk-storage middleware for all post types.
// This ensures req.body (including `type`) is always parsed before the handler runs.
const diskStorage = multer.diskStorage({
    destination: (req, file, cb) => {
        const dir = './temp/uploads';
        console.log(`[Multer] Receiving file: ${file.originalname} -> saving to ${dir}`);
        if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
        cb(null, dir);
    },
    filename: (req, file, cb) => cb(null, `${Date.now()}-${file.originalname}`)
});

const uploadFields = multer({ storage: diskStorage }).fields([
    { name: 'images', maxCount: 5 },
    { name: 'video',  maxCount: 1 }
]);

// ── Helpers ───────────────────────────────────────────────────────────────────
const ACCEPTED_VIDEO_MIME = ['video/mp4', 'video/quicktime', 'video/webm'];
const MAX_VIDEO_BYTES = 100 * 1024 * 1024; // 100 MB

function sanitizePostResponse(post) {
    if (!post) return null;
    if (post.deleted) {
        return {
            id: post.id,
            title: '[deleted]',
            community: post.community,
            deleted: true,
            createdAt: post.createdAt
        };
    }
    return post;
}

// ═════════════════════════════════════════════════════════════════════════════
// POST /posts — Create a post (text | image | link | video)
// ═════════════════════════════════════════════════════════════════════════════
router.post('/posts', uploadFields, async (req, res) => {
    try {
        console.log(`[CreatePost] Received post request. Title: "${req.body?.title}"`);
        const { title, community, body, url, flair, nsfw, spoiler, oc } = req.body;

        // ── Validation ────────────────────────────────────────────────────
        if (!title || title.length < 1 || title.length > 300) {
            return res.status(400).json({ error: 'Title is required and must be 1–300 characters.' });
        }
        if (!community) {
            return res.status(400).json({ error: 'Community is required.' });
        }

        // Validate community exists — resolve user-service via Consul, fall back to api-gateway
        try {
            let userServiceUrl = await consulService.resolve('user');
            if (!userServiceUrl) {
                userServiceUrl = process.env.API_GATEWAY_URL || 'http://api-gateway:8088';
            }
            const commRes = await fetch(`${userServiceUrl}/communities/${community}`);
            if (!commRes.ok) {
                return res.status(400).json({ error: `Community r/${community} does not exist.` });
            }
        } catch (err) {
            console.error(`[CreatePost] Warning: Failed to validate community: ${err.message}`);
        }

        const postId = uuidv4();
        const authorId = req.headers['x-user-id'] || '';
        const author   = req.headers['x-username'] || '';

        if (authorId && await banService.isBanned(authorId, community)) {
            return res.status(403).json({ error: 'You are banned from this community.' });
        }

        const postData = {
            title,
            community,
            authorId,
            author,
            body:    body    || '',
            url:     url     || '',
            flair:   flair   || '',
            nsfw:    nsfw    === 'true' || nsfw    === true,
            spoiler: spoiler === 'true' || spoiler === true,
            oc:      oc      === 'true' || oc      === true,
            images:  [],
            video:   null
        };

        // ── Process Images if present ─────────────────────────────────────
        if (req.files?.images) {
            console.log(`[CreatePost] Processing ${req.files.images.length} images...`);
            imageService.validateGallery(req.files.images);
            postData.images = await imageService.processGallery(req.files.images, postId);
            console.log(`[CreatePost] Image processing complete.`);
        }

        // ── Process Video if present ──────────────────────────────────────
        const videoFile = req.files?.video?.[0];
        let s3Key = null;

        if (videoFile) {
            if (!ACCEPTED_VIDEO_MIME.includes(videoFile.mimetype)) {
                return res.status(400).json({ error: 'Unsupported video format. Accepted: MP4, MOV, WebM.' });
            }
            if (videoFile.size > MAX_VIDEO_BYTES) {
                return res.status(400).json({ error: 'Video exceeds 100 MB limit.' });
            }
            
            s3Key = `${postId}.mp4`;
            postData.video = { status: 'UPLOADING', s3Key };
            console.log(`[CreatePost] Video file detected: ${videoFile.originalname}`);
        }

        console.log(`[CreatePost] Saving post to database...`);
        const post = await PostModel.create(postId, postData);
        console.log(`[CreatePost] Post saved. ID: ${postId}`);

        // ── Publish post.created event for feed-service ───────────────────
        const postType = videoFile ? 'video' : (req.files?.images?.length ? 'image' : (url ? 'link' : 'text'));
        await kafkaService.publish('post', {
            id:           postId,
            title:        post.title,
            body:         post.body       || '',
            community:    post.community,
            authorId:     post.authorId,
            author:       post.author,
            type:         postType,
            upvotes:      0,
            downvotes:    0,
            commentCount: 0,
            createdAt:    post.createdAt,
            images:       post.images || [],
            video:        post.video ? { status: post.video.status, playbackUrl: post.video.playbackUrl || '' } : null,
        });

        if (videoFile && s3Key) {
            res.status(202).json({
                postId,
                status: 'uploading',
                message: 'Post created. Video upload started in background.',
                post: sanitizePostResponse(post)
            });

            const uploadTask = new Upload({
                client: storageService.client,
                params: {
                    Bucket: process.env.S3_STAGING_BUCKET || 'staging',
                    Key: s3Key,
                    Body: fs.createReadStream(videoFile.path),
                    ContentType: videoFile.mimetype,
                },
            });

            uploadTask.done().then(async () => {
                await PostModel.update(postId, { 'video.status': 'UPLOADED' });
                await kafkaService.publish('video.uploaded', { postId, s3Key });
                fs.unlinkSync(videoFile.path);
            }).catch(async (err) => {
                console.error('[VideoUpload] Error:', err.message);
                await PostModel.update(postId, { 'video.status': 'FAILED' });
                if (fs.existsSync(videoFile.path)) fs.unlinkSync(videoFile.path);
            });
        } else {
            // No video, return 201 Created
            res.status(201).json(sanitizePostResponse(post));
        }

    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// ═════════════════════════════════════════════════════════════════════════════
// GET /posts — List all posts (with filtering & pagination)
// ═════════════════════════════════════════════════════════════════════════════
router.get('/posts', async (req, res) => {
    try {
        const { community, author, authorId, dateRange, limit, page } = req.query;

        const result = await PostModel.findList({
            community,
            author,
            authorId,
            dateRange,
            limit: parseInt(limit) || 20,
            page:  parseInt(page) || 1
        });

        res.json({
            ...result,
            posts: result.posts.map(sanitizePostResponse)
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// ═════════════════════════════════════════════════════════════════════════════
// GET /posts/:id — Get a post by ID
// ═════════════════════════════════════════════════════════════════════════════
router.get('/posts/:id', async (req, res) => {
    try {
        const post = await PostModel.findById(req.params.id);
        if (!post) return res.status(404).json({ error: 'Post not found.' });
        res.json(sanitizePostResponse(post));
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// ═════════════════════════════════════════════════════════════════════════════
// GET /posts/:id/status — Poll video upload/transcoding status
// ═════════════════════════════════════════════════════════════════════════════
router.get('/posts/:id/status', async (req, res) => {
    try {
        const post = await PostModel.findById(req.params.id);
        if (!post) return res.status(404).json({ error: 'Post not found.' });
        
        if (!post.video) {
            return res.status(400).json({ error: 'Status polling is only available for posts containing video.' });
        }
        res.json({
            postId:      post.id,
            status:      post.video?.status?.toLowerCase() || 'unknown',
            resolutions: post.video?.resolutions || [],
            playbackUrl: post.video?.playbackUrl || null,
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// ═════════════════════════════════════════════════════════════════════════════
// GET /posts/:id/history — Get full versioned edit history (text & link only)
// ═════════════════════════════════════════════════════════════════════════════
router.get('/posts/:id/history', async (req, res) => {
    try {
        const post = await PostModel.findById(req.params.id);
        if (!post) return res.status(404).json({ error: 'Post not found.' });
        
        res.json({
            postId:       post.id,
            currentVersion: post.editVersion,
            history:      post.editHistory || []
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// ═════════════════════════════════════════════════════════════════════════════
// PATCH /posts/:id — Edit a post
// ═════════════════════════════════════════════════════════════════════════════
router.patch('/posts/:id', async (req, res) => {
    try {
        const post = await PostModel.findById(req.params.id);
        if (!post) return res.status(404).json({ error: 'Post not found.' });
        if (post.deleted) return res.status(410).json({ error: 'Post has been deleted.' });

        const { title, flair, body, url } = req.body;
        const updates = {};

        // All fields can be edited if present
        if (title !== undefined) {
            if (title.length < 1 || title.length > 300) {
                return res.status(400).json({ error: 'Title must be 1–300 characters.' });
            }
            updates.title = title;
        }
        if (flair !== undefined) updates.flair = flair;
        if (body  !== undefined) updates.body  = body;
        if (url   !== undefined) updates.url   = url;

        // Note: Media (images/video) remains locked after submission for consistency

        // Append a versioned snapshot to editHistory
        const newVersion = (post.editVersion || 0) + 1;
        updates.editVersion = newVersion;
        updates.$push = {
            editHistory: {
                version:  newVersion,
                title:    post.title,
                body:     post.body,
                url:      post.url,
                flair:    post.flair,
                editedAt: new Date()
            }
        };

        // Mongoose doesn't support $push via findOneAndUpdate through our helper directly —
        // call Mongoose directly for the $push case
        const { $push, ...regularUpdates } = updates;
        let updatedPost;
        if ($push) {
            updatedPost = await mongoose.model('Post').findOneAndUpdate(
                { id: req.params.id },
                { $set: { ...regularUpdates, updatedAt: new Date() }, $push },
                { new: true }
            );
        } else {
            updatedPost = await PostModel.update(req.params.id, regularUpdates);
        }

        res.json(updatedPost);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// ═════════════════════════════════════════════════════════════════════════════
// DELETE /posts/:id — Soft-delete a post
// ═════════════════════════════════════════════════════════════════════════════
router.delete('/posts/:id', async (req, res) => {
    try {
        const post = await PostModel.findById(req.params.id);
        if (!post) return res.status(404).json({ error: 'Post not found.' });
        if (post.deleted) return res.status(410).json({ error: 'Post already deleted.' });

        await PostModel.update(req.params.id, {
            deleted:   true,
            deletedAt: new Date(),
        });

        // ── Publish post.deleted event for search service ──────────────────
        await kafkaService.publish('post.deleted', { id: req.params.id });

        res.json({ message: 'Post deleted. Media will be purged from storage within 24 hours.' });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

router.post('/posts/:id/vote', async (req, res) => {
    try {
        const { direction } = req.body; // 1 upvote | -1 downvote | 0 clear
        const postId = req.params.id;
        const userId = req.headers['x-user-id'];

        if (![1, -1, 0].includes(direction)) {
            return res.status(400).json({ error: 'Invalid vote direction. Use 1, -1, or 0.' });
        }
        const post = await PostModel.findById(postId);
        if (!post) return res.status(404).json({ error: 'Post not found.' });

        let inc = {};
        let mongoUpdate = {};

        if (userId) {
            // Per-user tracked voting (prevents double votes)
            const prev = post.userVotes?.get(userId) ?? 0;
            if (prev === direction) {
                return res.json({
                    id: post.id,
                    upvotes: post.upvotes,
                    downvotes: post.downvotes,
                    score: post.upvotes - post.downvotes,
                    userVote: prev,
                });
            }
            // Undo previous
            if (prev === 1)  inc.upvotes   = (inc.upvotes   || 0) - 1;
            if (prev === -1) inc.downvotes  = (inc.downvotes  || 0) - 1;
            // Apply new
            if (direction === 1)  inc.upvotes   = (inc.upvotes   || 0) + 1;
            if (direction === -1) inc.downvotes  = (inc.downvotes  || 0) + 1;

            mongoUpdate = { $inc: inc };
            if (direction === 0) {
                mongoUpdate.$unset = { [`userVotes.${userId}`]: '' };
            } else {
                mongoUpdate.$set = { [`userVotes.${userId}`]: direction };
            }
        } else {
            // Fallback: simple increment without per-user tracking
            if (direction === 1)  mongoUpdate.$inc = { upvotes: 1 };
            if (direction === -1) mongoUpdate.$inc = { downvotes: 1 };
        }

        if (!mongoUpdate.$inc && !mongoUpdate.$set && !mongoUpdate.$unset) {
            return res.json({ id: post.id, upvotes: post.upvotes, downvotes: post.downvotes, score: post.upvotes - post.downvotes, userVote: 0 });
        }

        const updatedPost = await mongoose.model('Post').findOneAndUpdate(
            { id: postId },
            mongoUpdate,
            { new: true }
        );

        await kafkaService.publish('post', {
            id:           updatedPost.id,
            title:        updatedPost.title,
            body:         updatedPost.body,
            community:    updatedPost.community,
            authorId:     updatedPost.authorId,
            author:       updatedPost.author,
            type:         updatedPost.video ? 'video' : (updatedPost.images?.length ? 'image' : 'text'),
            upvotes:      updatedPost.upvotes,
            downvotes:    updatedPost.downvotes,
            commentCount: updatedPost.commentCount,
            createdAt:    updatedPost.createdAt,
            images:       updatedPost.images || [],
            video:        updatedPost.video ? { status: updatedPost.video.status, playbackUrl: updatedPost.video.playbackUrl || '' } : null,
        });

        res.json({
            id: updatedPost.id,
            upvotes: updatedPost.upvotes,
            downvotes: updatedPost.downvotes,
            score: updatedPost.upvotes - updatedPost.downvotes,
            userVote: direction,
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

module.exports = router;
