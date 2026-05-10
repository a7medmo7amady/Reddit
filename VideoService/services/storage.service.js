const { S3Client, PutObjectCommand, GetObjectCommand, CreateBucketCommand, HeadBucketCommand, DeleteObjectCommand, ListObjectsV2Command } = require('@aws-sdk/client-s3');
const { Upload } = require('@aws-sdk/lib-storage');
const fs = require('fs');
const path = require('path');

class StorageService {
    constructor() {
        this.client = new S3Client({
            endpoint: process.env.S3_ENDPOINT || 'http://localhost:9000',
            region: process.env.S3_REGION || 'us-east-1',
            credentials: {
                accessKeyId: process.env.S3_ACCESS_KEY || 'minioadmin',
                secretAccessKey: process.env.S3_SECRET_KEY || 'minioadmin',
            },
            forcePathStyle: true,
            requestTimeout: 10000, // Prevent indefinite hangs
        });
        this.bucket = process.env.S3_BUCKET || 'videos';
    }

    async initialize() {
        const buckets = [
            process.env.S3_STAGING_BUCKET || 'staging',
            process.env.S3_SERVING_BUCKET || 'serving',
            process.env.S3_IMAGES_BUCKET  || 'images',
        ];

        for (const bucket of buckets) {
            try {
                await this.client.send(new HeadBucketCommand({ Bucket: bucket }));
            } catch (error) {
                if (error.name === 'NotFound' || error.$metadata?.httpStatusCode === 404) {
                    await this.client.send(new CreateBucketCommand({ Bucket: bucket }));
                } else {
                    throw error;
                }
            }
        }
    }

    /** Upload a file from disk to S3. */
    async uploadFile(localPath, s3Key, bucketName) {
        const targetBucket = bucketName || process.env.S3_SERVING_BUCKET || 'serving';
        const upload = new Upload({
            client: this.client,
            params: {
                Bucket: targetBucket,
                Key: s3Key,
                Body: fs.createReadStream(localPath),
                ContentType: this.getContentType(s3Key),
            },
        });
        return upload.done();
    }

    /** Upload an in-memory Buffer to S3 (used for Sharp-processed images). */
    async uploadBuffer(buffer, s3Key, contentType, bucketName) {
        const targetBucket = bucketName || process.env.S3_SERVING_BUCKET || 'serving';
        const command = new PutObjectCommand({
            Bucket: targetBucket,
            Key: s3Key,
            Body: buffer,
            ContentType: contentType || this.getContentType(s3Key),
            ContentLength: buffer.length,
        });
        return this.client.send(command);
    }

    /** Download a file from S3 to disk. */
    async downloadFile(s3Key, localPath, bucketName) {
        const targetBucket = bucketName || process.env.S3_STAGING_BUCKET || 'staging';
        const command = new GetObjectCommand({ Bucket: targetBucket, Key: s3Key });
        const response = await this.client.send(command);
        const fileStream = fs.createWriteStream(localPath);
        return new Promise((resolve, reject) => {
            response.Body.pipe(fileStream).on('error', reject).on('finish', resolve);
        });
    }

    /** Stream a file from S3 directly to the HTTP response. */
    async streamFile(s3Key, bucketName) {
        const targetBucket = bucketName || process.env.S3_SERVING_BUCKET || 'serving';
        const command = new GetObjectCommand({ Bucket: targetBucket, Key: s3Key });
        const response = await this.client.send(command);
        return {
            stream: response.Body,
            contentType: response.ContentType,
            contentLength: response.ContentLength
        };
    }

    /** Delete a single object from S3. */
    async deleteFile(s3Key, bucketName) {
        const targetBucket = bucketName || process.env.S3_SERVING_BUCKET || 'serving';
        const command = new DeleteObjectCommand({ Bucket: targetBucket, Key: s3Key });
        return this.client.send(command);
    }

    /**
     * Delete all objects under a key prefix (simulates folder deletion).
     * Used by the purge service for HLS segment cleanup.
     */
    async deleteFolder(prefix, bucketName) {
        const targetBucket = bucketName || process.env.S3_SERVING_BUCKET || 'serving';

        const listed = await this.client.send(new ListObjectsV2Command({
            Bucket: targetBucket,
            Prefix: prefix,
        }));

        if (!listed.Contents || listed.Contents.length === 0) return;

        for (const obj of listed.Contents) {
            await this.client.send(new DeleteObjectCommand({
                Bucket: targetBucket,
                Key: obj.Key,
            }));
        }
    }

    getContentType(filename) {
        const ext = path.extname(filename).toLowerCase();
        switch (ext) {
            case '.m3u8': return 'application/x-mpegURL';
            case '.ts':   return 'video/MP2T';
            case '.mp4':  return 'video/mp4';
            case '.webp': return 'image/webp';
            case '.gif':  return 'image/gif';
            case '.jpg':
            case '.jpeg': return 'image/jpeg';
            case '.png':  return 'image/png';
            default:      return 'application/octet-stream';
        }
    }
}

module.exports = new StorageService();
