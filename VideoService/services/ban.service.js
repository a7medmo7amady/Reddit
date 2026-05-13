const Redis = require('ioredis');

let client = null;

function getClient() {
    if (!client) {
        client = new Redis(process.env.REDIS_URL || 'redis://localhost:6379', {
            lazyConnect: true,
            enableReadyCheck: false,
        });
        client.on('error', (err) => console.error('[BanService] Redis error:', err.message));
    }
    return client;
}

// Mirrors feed-service ban.go key schema: ban:{userId}:{community}
async function isBanned(userId, community) {
    if (!userId || !community) return false;
    try {
        const n = await getClient().exists(`ban:${userId}:${community}`);
        return n > 0;
    } catch (err) {
        console.error('[BanService] isBanned check failed:', err.message);
        return false;
    }
}

module.exports = { isBanned };
