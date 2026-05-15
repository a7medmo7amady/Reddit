const express = require('express');
const router = express.Router();
const PostModel = require('../models/post.model');
const CommentModel = require('../models/comment.model');

router.get('/search', async (req, res) => {
    try {
        const { q, limit = 20, page = 1 } = req.query;
        if (!q || q.trim().length === 0) {
            return res.status(400).json({ error: 'Query parameter q is required.' });
        }

        const searchRegex = new RegExp(q.trim(), 'i');
        const l = parseInt(limit) || 20;
        const p = parseInt(page) || 1;
        const skip = (p - 1) * l;

        // Search posts by title, body, community, or author
        const postQuery = {
            deleted: false,
            $or: [
                { title: { $regex: searchRegex } },
                { body: { $regex: searchRegex } },
                { community: { $regex: searchRegex } },
                { author: { $regex: searchRegex } }
            ]
        };

        const posts = await PostModel.findList({
            limit: l,
            page: p,
            // We need to override findList's query building; call mongoose directly for search
        });

        // Actually, PostModel.findList doesn't support arbitrary queries.
        // Call mongoose directly for search flexibility.
        const Post = require('mongoose').model('Post');
        const Comment = require('mongoose').model('Comment');

        const [postResults, postTotal] = await Promise.all([
            Post.find(postQuery).sort({ createdAt: -1 }).skip(skip).limit(l),
            Post.countDocuments(postQuery)
        ]);

        // Search comments by body or author
        const commentQuery = {
            deleted: false,
            $or: [
                { body: { $regex: searchRegex } },
                { author: { $regex: searchRegex } }
            ]
        };

        const [commentResults, commentTotal] = await Promise.all([
            Comment.find(commentQuery).sort({ createdAt: -1 }).skip(skip).limit(l),
            Comment.countDocuments(commentQuery)
        ]);

        res.json({
            posts: {
                items: postResults,
                total: postTotal,
                page: p,
                limit: l,
                pages: Math.ceil(postTotal / l)
            },
            comments: {
                items: commentResults,
                total: commentTotal,
                page: p,
                limit: l,
                pages: Math.ceil(commentTotal / l)
            },
            query: q
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

module.exports = router;
