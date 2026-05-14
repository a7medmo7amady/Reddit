const PostModel = require('../models/post.model');
const kafkaService = require('./kafka.service');

const SEED_POSTS = [
    { id: 'seed-post-001', community: 'programming', author: 'torvalds',          upvotes: 42, commentCount: 31, createdAt: '2025-05-12T08:00:00Z', title: "I just realized I've been writing the same bug for 20 years",              body: "It's always an off-by-one error. Always." },
    { id: 'seed-post-002', community: 'golang',       author: 'rob_pike',          upvotes: 38, commentCount: 24, createdAt: '2025-05-12T09:10:00Z', title: 'Go 1.25 is out — generics just got faster',                               body: 'Compile times are down 30% in the new release. Full notes in the link.' },
    { id: 'seed-post-003', community: 'worldnews',    author: 'newsbot',           upvotes: 35, commentCount: 89, createdAt: '2025-05-12T07:30:00Z', title: 'Scientists confirm fastest internet speed record: 402 Tbps',              body: 'Researchers at UCL broke the record using a new multi-band fiber optic technique.' },
    { id: 'seed-post-004', community: 'gaming',       author: 'xXslayer99Xx',      upvotes: 31, commentCount: 56, createdAt: '2025-05-11T22:00:00Z', title: 'After 3,000 hours I finally beat the final boss without taking damage',   body: 'No summons, no shields, no cheese. Pure skill.' },
    { id: 'seed-post-005', community: 'science',      author: 'dr_cosmos',         upvotes: 28, commentCount: 18, createdAt: '2025-05-12T06:00:00Z', title: 'NASA confirms water ice found in permanently shadowed craters on the Moon', body: 'The findings significantly boost the case for a sustainable lunar base.' },
    { id: 'seed-post-006', community: 'technology',   author: 'hackernews_mirror', upvotes: 25, commentCount: 22, createdAt: '2025-05-11T18:00:00Z', title: "Chrome now uses 40% less RAM — here's how they did it",                  body: 'A deep dive into the new memory management system shipped in Chrome 130.' },
    { id: 'seed-post-007', community: 'AskReddit',    author: 'curious_carl',      upvotes: 21, commentCount: 143, createdAt: '2025-05-12T10:00:00Z', title: "What's a skill you learned during COVID that you still use every day?",  body: '' },
    { id: 'seed-post-008', community: 'python',       author: 'guido_fan',         upvotes: 18, commentCount: 19, createdAt: '2025-05-11T14:00:00Z', title: 'Python 3.14 drops the GIL by default — what this means for your code',  body: 'Free-threaded Python is now the default build.' },
    { id: 'seed-post-009', community: 'linux',        author: 'arch_enjoyer',      upvotes: 14, commentCount: 34, createdAt: '2025-05-11T20:00:00Z', title: 'Wayland finally works perfectly on my setup after 4 years of trying',    body: 'Screen sharing, gaming, dual monitors — all working.' },
    { id: 'seed-post-010', community: 'webdev',       author: 'css_pain',          upvotes: 9,  commentCount: 39, createdAt: '2025-05-10T22:00:00Z', title: 'CSS is finally getting native masonry layout in 2025',                   body: 'No more JavaScript hacks for Pinterest-style grids.' },
];

async function seed() {
    let seeded = 0;
    for (const p of SEED_POSTS) {
        const existing = await PostModel.findById(p.id);
        if (existing) continue;

        const post = await PostModel.create(p.id, {
            title:        p.title,
            body:         p.body || '',
            community:    p.community,
            author:       p.author,
            authorId:     p.author,
            upvotes:      p.upvotes,
            downvotes:    0,
            commentCount: p.commentCount,
            createdAt:    new Date(p.createdAt),
        });

        await kafkaService.publish('post', {
            id:           post.id,
            title:        post.title,
            body:         post.body,
            community:    post.community,
            authorId:     post.authorId,
            author:       post.author,
            type:         'text',
            upvotes:      post.upvotes,
            downvotes:    post.downvotes,
            commentCount: post.commentCount,
            createdAt:    post.createdAt,
        });

        seeded++;
    }
    if (seeded > 0) {
        console.log(`[Seed] Inserted ${seeded} seed post(s) into MongoDB.`);
    } else {
        console.log('[Seed] All seed posts already exist, skipping.');
    }
}

module.exports = { seed };
