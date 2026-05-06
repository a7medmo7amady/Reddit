const express = require('express');
const router = express.Router();
const { v4: uuidv4 } = require('uuid');
const multer = require('multer');
const fs = require('fs');
const path = require('path');
const VideoModel = require('../models/video.model');
const kafkaService = require('../services/kafka.service');
const storageService = require('../services/storage.service');
const { Upload } = require('@aws-sdk/lib-storage');

const storage = multer.diskStorage({
    destination: (req, file, cb) => {
        const dir = './temp/uploads';
        if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
        cb(null, dir);
    },
    filename: (req, file, cb) => {
        cb(null, `${Date.now()}-${file.originalname}`);
    }
});

const upload = multer({ storage });

router.post('/videos', upload.single('video'), async (req, res) => {
    try {
        const { title, description } = req.body;
        const file = req.file;
        if (!file) return res.status(400).json({ error: 'No video file provided' });

        const videoId = uuidv4();
        const s3Key = `${videoId}.mp4`;

        await VideoModel.create(videoId, {
            title: title || 'Untitled',
            description: description || '',
            status: 'UPLOADING'
        });

        res.status(202).json({ videoId, status: 'uploading' });

        const uploadTask = new Upload({
            client: storageService.client,
            params: {
                Bucket: process.env.S3_STAGING_BUCKET || 'staging',
                Key: s3Key,
                Body: fs.createReadStream(file.path),
                ContentType: file.mimetype,
            },
        });

        uploadTask.done().then(async () => {
            await VideoModel.update(videoId, { status: 'UPLOADED' });
            await kafkaService.publish('video.uploaded', { videoId, s3Key });
            fs.unlinkSync(file.path);
        }).catch(async () => {
            await VideoModel.update(videoId, { status: 'FAILED' });
            if (fs.existsSync(file.path)) fs.unlinkSync(file.path);
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

router.get('/videos/:id/status', async (req, res) => {
    try {
        const video = await VideoModel.findById(req.params.id);
        if (!video) return res.status(404).json({ error: 'Not found' });
        res.json({ status: video.status.toLowerCase() });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

router.get('/videos/:id', async (req, res) => {
    try {
        const video = await VideoModel.findById(req.params.id);
        if (!video) return res.status(404).json({ error: 'Not found' });
        res.json({
            videoId: video.id,
            status: video.status.toLowerCase(),
            duration: video.duration || 120,
            thumbnailUrl: video.thumbnailUrl || '',
            previewUrl: video.previewUrl || '',
            manifestUrl: video.playbackUrl || '',
            resolutions: video.resolutions || ['360p', '720p']
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

router.get('/video/:id/:filename', async (req, res) => {
    try {
        const { id, filename } = req.params;
        const { stream, contentType } = await storageService.streamFile(`video/${id}/${filename}`);
        res.set('Content-Type', contentType);
        stream.pipe(res);
    } catch (error) {
        res.status(404).json({ error: 'Asset not found' });
    }
});

router.get('/thumbs/:filename', async (req, res) => {
    try {
        const { stream, contentType } = await storageService.streamFile(`thumbs/${req.params.filename}`);
        res.set('Content-Type', contentType);
        stream.pipe(res);
    } catch (error) {
        res.status(404).json({ error: 'Thumbnail not found' });
    }
});

router.get('/previews/:filename', async (req, res) => {
    try {
        const { stream, contentType } = await storageService.streamFile(`previews/${req.params.filename}`);
        res.set('Content-Type', contentType);
        stream.pipe(res);
    } catch (error) {
        res.status(404).json({ error: 'Preview not found' });
    }
});

module.exports = router;
