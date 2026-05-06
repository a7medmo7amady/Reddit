const { S3Client, PutObjectCommand, GetObjectCommand, CreateBucketCommand, HeadBucketCommand } = require('@aws-sdk/client-s3');
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
        });
        this.bucket = process.env.S3_BUCKET || 'videos';
    }

    async initialize() {
        const buckets = [
            process.env.S3_STAGING_BUCKET || 'staging',
            process.env.S3_SERVING_BUCKET || 'serving'
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

    async downloadFile(s3Key, localPath, bucketName) {
        const targetBucket = bucketName || process.env.S3_STAGING_BUCKET || 'staging';
        const command = new GetObjectCommand({ Bucket: targetBucket, Key: s3Key });
        const response = await this.client.send(command);
        const fileStream = fs.createWriteStream(localPath);
        return new Promise((resolve, reject) => {
            response.Body.pipe(fileStream).on('error', reject).on('finish', resolve);
        });
    }

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

    getContentType(filename) {
        const ext = path.extname(filename).toLowerCase();
        switch (ext) {
            case '.m3u8': return 'application/x-mpegURL';
            case '.ts': return 'video/MP2T';
            case '.mp4': return 'video/mp4';
            case '.jpg':
            case '.jpeg': return 'image/jpeg';
            case '.png': return 'image/png';
            default: return 'application/octet-stream';
        }
    }
}

module.exports = new StorageService();
