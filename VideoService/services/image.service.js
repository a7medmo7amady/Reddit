const sharp = require('sharp');
const path = require('path');
const fs = require('fs');
const { v4: uuidv4 } = require('uuid');
const storageService = require('./storage.service');

// Accepted MIME types per FR-UC-02
const ACCEPTED_MIME_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
const MAX_FILE_SIZE_BYTES = 20 * 1024 * 1024; // 20 MB
const MAX_GALLERY_IMAGES = 5;

// Breakpoints: { variant, width }  — height scales proportionally
const BREAKPOINTS = [
    { variant: 'thumbnail', width: 140 },
    { variant: 'preview',   width: 640 },
    { variant: 'full',      width: 1080 }
];

class ImageService {
    /**
     * Validate a single file upload.
     * @param {Express.Multer.File} file
     * @throws if file type or size is invalid
     */
    validate(file) {
        if (!ACCEPTED_MIME_TYPES.includes(file.mimetype)) {
            throw new Error(`Unsupported image type: ${file.mimetype}. Accepted: JPEG, PNG, GIF, WebP.`);
        }
        if (file.size > MAX_FILE_SIZE_BYTES) {
            throw new Error(`Image exceeds 20 MB limit (got ${(file.size / 1024 / 1024).toFixed(1)} MB).`);
        }
    }

    /**
     * Validate a gallery upload (up to 5 images).
     * @param {Express.Multer.File[]} files
     * @throws if count or individual file is invalid
     */
    validateGallery(files) {
        if (!files || files.length === 0) throw new Error('No image files provided.');
        if (files.length > MAX_GALLERY_IMAGES) {
            throw new Error(`Gallery cannot exceed ${MAX_GALLERY_IMAGES} images (got ${files.length}).`);
        }
        files.forEach(f => this.validate(f));
    }

    /**
     * Process a single image:
     *  1. Strip EXIF metadata
     *  2. Convert to WebP
     *  3. Resize to three breakpoints
     *  4. Upload all variants to the images bucket
     *
     * @param {string}  filePath    Path to the temp file on disk
     * @param {string}  postId      Used to namespace S3 keys
     * @param {string}  originalName
     * @returns {Promise<{ thumbnail: string, preview: string, full: string, originalName: string }>}
     */
    async processImage(filePath, postId, originalName) {
        const imageId = uuidv4();
        const urls = {};

        for (const bp of BREAKPOINTS) {
            const processedBuffer = await sharp(filePath)
                .rotate()               // auto-orient based on EXIF, then strip it
                .resize({ width: bp.width, withoutEnlargement: true })
                .webp({ quality: 82 })
                .toBuffer();

            const s3Key = `images/${postId}/${imageId}-${bp.variant}.webp`;

            await storageService.uploadBuffer(
                processedBuffer,
                s3Key,
                'image/webp',
                process.env.S3_IMAGES_BUCKET || 'images'
            );

            urls[bp.variant] = `/assets/images/${postId}/${imageId}-${bp.variant}.webp`;
        }

        return { ...urls, originalName };
    }

    /**
     * Process all images in a gallery upload.
     * Reads each file from disk, processes it, then deletes the temp file.
     * @param {Express.Multer.File[]} files  (disk storage — has file.path)
     * @param {string} postId
     * @returns {Promise<Array<{ thumbnail, preview, full, originalName }>>}
     */
    async processGallery(files, postId) {
        const results = [];
        for (let i = 0; i < files.length; i++) {
            const file = files[i];
            console.log(`  [ImageService] Processing file ${i+1}/${files.length}: ${file.originalname}`);
            const result = await this.processImage(file.path, postId, file.originalname || file.originalName);
            // Clean up temp file from disk after processing
            if (fs.existsSync(file.path)) fs.unlinkSync(file.path);
            results.push(result);
        }
        return results;
    }

    /**
     * Delete all image S3 keys for a post.
     * The keys are inferred from the stored URL paths.
     * @param {Array<{ thumbnail, preview, full }>} images
     */
    async deleteImages(images) {
        const bucket = process.env.S3_IMAGES_BUCKET || 'images';
        for (const img of images) {
            const keys = [img.thumbnail, img.preview, img.full]
                .filter(Boolean)
                .map(url => url.replace('/assets/images/', 'images/'));

            for (const key of keys) {
                await storageService.deleteFile(key, bucket).catch(() => {});
            }
        }
    }
}

module.exports = new ImageService();
