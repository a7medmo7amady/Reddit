const mongoose = require('mongoose');

const commentSchema = new mongoose.Schema({
    id:       { type: String, required: true, unique: true },
    postId:   { type: String, required: true, index: true },
    authorId: { type: String, required: true },
    author:   { type: String, default: '' },
    body:     { type: String, required: true, maxlength: 10000 },
    
    // For nested comments/replies
    parentId: { type: String, default: null, index: true },
    
    // Interaction metrics
    upvotes:   { type: Number, default: 0 },
    downvotes: { type: Number, default: 0 },
    
    deleted:   { type: Boolean, default: false },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now }
});

const Comment = mongoose.model('Comment', commentSchema);

class CommentModel {
    static async create(id, data) {
        const comment = new Comment({ id, ...data });
        return await comment.save();
    }

    static async findById(id) {
        return await Comment.findOne({ id });
    }

    static async findByPost(postId, { limit = 20, page = 1, parentId = null }) {
        const query = { postId, deleted: false, parentId };
        
        const skip = (page - 1) * limit;
        const comments = await Comment.find(query)
            .sort({ createdAt: -1 }) // Newest first
            .skip(skip)
            .limit(limit);
            
        const total = await Comment.countDocuments(query);
        
        return {
            comments,
            pagination: {
                total,
                page,
                limit,
                pages: Math.ceil(total / limit)
            }
        };
    }

    static async findList({ postId, author, authorId, limit = 20, page = 1, parentId = null }) {
        const query = { deleted: false };
        if (postId) query.postId = postId;
        if (author) query.author = author;
        if (authorId) query.authorId = authorId;
        if (parentId !== undefined) query.parentId = parentId;

        const skip = (page - 1) * limit;
        
        const comments = await Comment.find(query)
            .sort({ createdAt: -1 })
            .skip(skip)
            .limit(limit);

        const total = await Comment.countDocuments(query);

        return {
            comments,
            pagination: {
                total,
                page,
                limit,
                pages: Math.ceil(total / limit)
            }
        };
    }

    static async update(id, updates) {
        return await Comment.findOneAndUpdate(
            { id },
            { ...updates, updatedAt: Date.now() },
            { new: true }
        );
    }
}

module.exports = CommentModel;
