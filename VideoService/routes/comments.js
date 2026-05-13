const express = require('express');
const router = express.Router();
const { v4: uuidv4 } = require('uuid');
const CommentModel = require('../models/comment.model');
const PostModel = require('../models/post.model');

// ── GET /posts/:postId/comments ──────────────────────────────────────────────
// Fetch paginated comments for a specific post
router.get('/posts/:postId/comments', async (req, res) => {
    try {
        const { postId } = req.params;
        const limit = parseInt(req.query.limit) || 20;
        const page = parseInt(req.query.page) || 1;
        const parentId = req.query.parentId || null; // For fetching replies

        const result = await CommentModel.findByPost(postId, { limit, page, parentId });
        
        res.json(result);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// ── POST /posts/:postId/comments ─────────────────────────────────────────────
// Create a new comment on a post
router.post('/posts/:postId/comments', async (req, res) => {
    try {
        const { postId } = req.params;
        const { body, parentId } = req.body;
        const authorId = req.headers['x-user-id'] || 'anonymous';

        if (!body || body.trim().length === 0) {
            return res.status(400).json({ error: 'Comment body cannot be empty.' });
        }

        // Verify post exists
        const post = await PostModel.findById(postId);
        if (!post) {
            return res.status(404).json({ error: 'Post not found.' });
        }

        const commentId = uuidv4();
        const comment = await CommentModel.create(commentId, {
            postId,
            authorId,
            body,
            parentId: parentId || null
        });

        // Increment comment count on the post
        await PostModel.update(postId, { $inc: { commentCount: 1 } });

        res.status(201).json(comment);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// ── POST /comments/:id/vote ──────────────────────────────────────────────────
// Upvote or downvote a comment
router.post('/comments/:id/vote', async (req, res) => {
    try {
        const { direction } = req.body;
        const commentId = req.params.id;

        if (![1, -1, 0].includes(direction)) {
            return res.status(400).json({ error: 'Invalid vote direction.' });
        }

        const comment = await CommentModel.findById(commentId);
        if (!comment) return res.status(404).json({ error: 'Comment not found.' });

        const update = {};
        if (direction === 1) update.$inc = { upvotes: 1 };
        else if (direction === -1) update.$inc = { downvotes: 1 };

        const updatedComment = await CommentModel.update(commentId, update);

        res.json({
            id: updatedComment.id,
            upvotes: updatedComment.upvotes,
            downvotes: updatedComment.downvotes,
            score: updatedComment.upvotes - updatedComment.downvotes
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

module.exports = router;
