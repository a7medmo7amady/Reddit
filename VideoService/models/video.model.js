const videos = new Map();

class VideoModel {
    static async create(id, data) {
        const record = {
            id,
            status: 'PENDING',
            createdAt: new Date(),
            ...data
        };
        videos.set(id, record);
        return record;
    }

    static async findById(id) {
        return videos.get(id);
    }

    static async update(id, updates) {
        const video = videos.get(id);
        if (!video) return null;
        const updated = { ...video, ...updates, updatedAt: new Date() };
        videos.set(id, updated);
        return updated;
    }

    static async getAll() {
        return Array.from(videos.values());
    }
}

module.exports = VideoModel;
