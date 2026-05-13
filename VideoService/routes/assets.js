const express = require('express');
const router = express.Router();
const storageService = require('../services/storage.service');

// ── Video HLS segments & manifests ───────────────────────────────────────────
router.get('/assets/video/:id/:filename', async (req, res) => {
    try {
        const { id, filename } = req.params;
        const { stream, contentType } = await storageService.streamFile(
            `video/${id}/${filename}`,
            process.env.S3_SERVING_BUCKET || 'serving'
        );
        res.set('Content-Type', contentType);
        stream.pipe(res);
    } catch {
        res.status(404).json({ error: 'Asset not found.' });
    }
});

// ── Video thumbnails ──────────────────────────────────────────────────────────
router.get('/assets/thumbs/:filename', async (req, res) => {
    try {
        const { stream, contentType } = await storageService.streamFile(
            `thumbs/${req.params.filename}`,
            process.env.S3_SERVING_BUCKET || 'serving'
        );
        res.set('Content-Type', contentType);
        stream.pipe(res);
    } catch {
        res.status(404).json({ error: 'Thumbnail not found.' });
    }
});

// ── Video preview GIFs ────────────────────────────────────────────────────────
router.get('/assets/previews/:filename', async (req, res) => {
    try {
        const { stream, contentType } = await storageService.streamFile(
            `previews/${req.params.filename}`,
            process.env.S3_SERVING_BUCKET || 'serving'
        );
        res.set('Content-Type', contentType);
        stream.pipe(res);
    } catch {
        res.status(404).json({ error: 'Preview not found.' });
    }
});

// ── Processed images (thumbnail / preview / full WebP variants) ───────────────
// URL pattern: /assets/images/<postId>/<imageId>-<variant>.webp
router.get('/assets/images/:postId/:filename', async (req, res) => {
    try {
        const { postId, filename } = req.params;
        const { stream, contentType } = await storageService.streamFile(
            `images/${postId}/${filename}`,
            process.env.S3_IMAGES_BUCKET || 'images'
        );
        res.set('Content-Type', contentType);
        stream.pipe(res);
    } catch {
        res.status(404).json({ error: 'Image not found.' });
    }
});

module.exports = router;
