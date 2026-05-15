const mongoose = require('mongoose');

// ── Edit History Entry ────────────────────────────────────────────────────────
const editHistorySchema = new mongoose.Schema({
    version:     { type: Number, required: true },
    title:       { type: String },
    body:        { type: String },
    url:         { type: String },
    flair:       { type: String },
    editedAt:    { type: Date, default: Date.now }
}, { _id: false });

// ── Image Variant ─────────────────────────────────────────────────────────────
const imageVariantSchema = new mongoose.Schema({
    thumbnail:   { type: String, default: '' },  // 140px WebP
    preview:     { type: String, default: '' },  // 640px WebP
    full:        { type: String, default: '' },  // 1080px WebP
    originalName:{ type: String, default: '' }
}, { _id: false });

// ── Video Sub-document ────────────────────────────────────────────────────────
const videoSchema = new mongoose.Schema({
    status: {
        type: String,
        enum: ['PENDING', 'UPLOADING', 'UPLOADED', 'PROCESSING', 'READY', 'FAILED'],
        default: 'PENDING'
    },
    s3Key:        { type: String, default: '' },
    thumbnailUrl: { type: String, default: '' },
    previewUrl:   { type: String, default: '' },
    playbackUrl:  { type: String, default: '' },
    duration:     { type: Number, default: 0 },
    resolutions:  { type: [String], default: [] }
}, { _id: false });

// ── Post Schema ───────────────────────────────────────────────────────────────
const postSchema = new mongoose.Schema({
    id:          { type: String, required: true, unique: true },

    // Content fields — all optional to allow mixed content
    title:       { type: String, required: true, minlength: 1, maxlength: 300 },
    community:   { type: String, required: true },
    authorId:    { type: String, default: '' },
    author:      { type: String, default: '' },

    // Type-specific content (can now coexist)
    body:        { type: String, default: '' },
    url:         { type: String, default: '' },
    linkPreviewUrl: { type: String, default: '' },
    images:      { type: [imageVariantSchema], default: [] },
    video:       { type: videoSchema, default: null },

    // FR-UC-01 — optional flags
    flair:       { type: String, default: '' },
    nsfw:        { type: Boolean, default: false },
    spoiler:     { type: Boolean, default: false },
    oc:          { type: Boolean, default: false },

    // FR-UC-04 — edit history (text & link only)
    editVersion: { type: Number, default: 0 },
    editHistory: { type: [editHistorySchema], default: [] },

    // FR-UC-04 — soft-delete + purge tracking
    deleted:     { type: Boolean, default: false },
    deletedAt:   { type: Date, default: null },
    purged:      { type: Boolean, default: false },  // true once S3 files are cleaned up

    // Interaction metrics
    upvotes:      { type: Number, default: 0 },
    downvotes:    { type: Number, default: 0 },
    commentCount: { type: Number, default: 0 },
    // Per-user vote tracking: userId → 1 (up) | -1 (down)
    userVotes:    { type: Map, of: Number, default: {} },

    createdAt:   { type: Date, default: Date.now },
    updatedAt:   { type: Date, default: Date.now }
});

const Post = mongoose.model('Post', postSchema);

class PostModel {
    static async create(id, data) {
        const post = new Post({ id, ...data });
        return await post.save();
    }

    static async findById(id) {
        return await Post.findOne({ id });
    }

    static async update(id, updates) {
        return await Post.findOneAndUpdate(
            { id },
            { ...updates, updatedAt: Date.now() },
            { new: true }
        );
    }

    static async getAll() {
        return await Post.find({ deleted: false }).sort({ createdAt: -1 });
    }

    static async findList({ community, author, authorId, dateRange, limit = 10, page = 1 }) {
        const query = { deleted: false };
        if (community) query.community = community;
        if (author) query.author = author;
        if (authorId) query.authorId = authorId;

        if (dateRange && dateRange !== 'all') {
            const now = new Date();
            let startDate;
            if (dateRange === 'week') {
                startDate = new Date(now.setDate(now.getDate() - 7));
            } else if (dateRange === 'month') {
                startDate = new Date(now.setMonth(now.getMonth() - 1));
            }

            if (startDate) {
                query.createdAt = { $gte: startDate };
            }
        }

        const skip = (page - 1) * limit;
        
        const posts = await Post.find(query)
            .sort({ createdAt: -1 })
            .skip(skip)
            .limit(limit);

        const total = await Post.countDocuments(query);

        return {
            posts,
            pagination: {
                total,
                page,
                limit,
                pages: Math.ceil(total / limit)
            }
        };
    }

    /**
     * Returns all posts that have been deleted for more than 24 hours
     * and have not yet had their S3 assets purged.
     */
    static async findPurgeable() {
        const cutoff = new Date(Date.now() - 24 * 60 * 60 * 1000);
        return await Post.find({
            deleted: true,
            purged: false,
            deletedAt: { $lte: cutoff }
        });
    }
}

module.exports = PostModel;
