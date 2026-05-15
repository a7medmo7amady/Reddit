const express = require('express');
const router = express.Router();
const { GetObjectCommand, HeadObjectCommand } = require('@aws-sdk/client-s3');
const storageService = require('../services/storage.service');

function parseRange(header, total) {
    const m = header.match(/bytes=(\d+)-(\d*)/);
    if (!m) return null;
    const start = parseInt(m[1]);
    const end = m[2] ? parseInt(m[2]) : total - 1;
    return { start, end, length: end - start + 1 };
}

async function streamAsset(req, res, s3Key, bucket, filename) {
    try {
        // Get file size first
        const head = await storageService.client.send(
            new HeadObjectCommand({ Bucket: bucket, Key: s3Key })
        );
        const total = head.ContentLength;
        const contentType = head.ContentType || storageService.getContentType(s3Key);

        const rangeHeader = req.headers['range'];

        if (rangeHeader) {
            const range = parseRange(rangeHeader, total);
            if (!range || range.start >= total) {
                res.status(416).set('Content-Range', `bytes */${total}`).end();
                return;
            }

            const cmd = new GetObjectCommand({
                Bucket: bucket,
                Key: s3Key,
                Range: `bytes=${range.start}-${range.end}`,
            });
            const obj = await storageService.client.send(cmd);

            res.status(206).set({
                'Content-Type':   contentType,
                'Content-Length': range.length,
                'Content-Range':  `bytes ${range.start}-${range.end}/${total}`,
                'Accept-Ranges':  'bytes',
            });
            obj.Body.pipe(res);
        } else {
            const cmd = new GetObjectCommand({ Bucket: bucket, Key: s3Key });
            const obj = await storageService.client.send(cmd);

            const disposition = filename
                ? `attachment; filename="${filename}"`
                : 'inline';

            res.status(200).set({
                'Content-Type':        contentType,
                'Content-Length':      total,
                'Accept-Ranges':       'bytes',
                'Content-Disposition': disposition,
            });
            obj.Body.pipe(res);
        }
    } catch {
        res.status(404).json({ error: 'Asset not found.' });
    }
}

// ── Images ────────────────────────────────────────────────────────────────────
router.get('/assets/images/:postId/:filename', (req, res) => {
    const { postId, filename } = req.params;
    const bucket = process.env.S3_IMAGES_BUCKET || 'images';
    streamAsset(req, res, `images/${postId}/${filename}`, bucket, null);
});

// ── Video (mp4 / HLS segments) — range-request aware ─────────────────────────
router.get('/assets/video/:id/:filename', (req, res) => {
    const { id, filename } = req.params;
    const bucket = process.env.S3_SERVING_BUCKET || 'serving';
    streamAsset(req, res, `video/${id}/${filename}`, bucket, filename);
});

// ── Video thumbnails ──────────────────────────────────────────────────────────
router.get('/assets/thumbs/:filename', (req, res) => {
    const bucket = process.env.S3_SERVING_BUCKET || 'serving';
    streamAsset(req, res, `thumbs/${req.params.filename}`, bucket, null);
});

// ── Video preview GIFs ────────────────────────────────────────────────────────
router.get('/assets/previews/:filename', (req, res) => {
    const bucket = process.env.S3_SERVING_BUCKET || 'serving';
    streamAsset(req, res, `previews/${req.params.filename}`, bucket, null);
});

module.exports = router;
