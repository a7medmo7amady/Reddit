const mongoose = require('mongoose');

const videoSchema = new mongoose.Schema({
    id: { type: String, required: true, unique: true },
    title: { type: String, default: 'Untitled' },
    description: { type: String, default: '' },
    status: { 
        type: String, 
        enum: ['PENDING', 'UPLOADING', 'UPLOADED', 'PROCESSING', 'READY', 'FAILED'], 
        default: 'PENDING' 
    },
    thumbnailUrl: { type: String, default: '' },
    previewUrl: { type: String, default: '' },
    playbackUrl: { type: String, default: '' },
    duration: { type: Number, default: 0 },
    resolutions: { type: [String], default: [] },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now }
});

const Video = mongoose.model('Video', videoSchema);

class VideoModel {
    static async create(id, data) {
        const video = new Video({ id, ...data });
        return await video.save();
    }

    static async findById(id) {
        return await Video.findOne({ id });
    }

    static async update(id, updates) {
        return await Video.findOneAndUpdate(
            { id }, 
            { ...updates, updatedAt: Date.now() }, 
            { new: true }
        );
    }

    static async getAll() {
        return await Video.find().sort({ createdAt: -1 });
    }
}

module.exports = VideoModel;
