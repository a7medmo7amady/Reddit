const { Kafka } = require('kafkajs');

class KafkaService {
    constructor() {
        this.kafka = new Kafka({
            clientId: 'video-service',
            brokers: (process.env.KAFKA_BROKERS || 'localhost:9092').split(','),
            retry: { initialRetryTime: 5000, retries: 20 },
            connectionTimeout: 45000,
            requestTimeout: 120000,
        });
        this.producer = this.kafka.producer();
        this.consumer = this.kafka.consumer({ 
            groupId: 'video-processing-group',
            sessionTimeout: 300000,
            rebalanceTimeout: 60000,
            heartbeatInterval: 10000
        });
        this.isConnected = false;
        this.callbacks = {};
    }

    async connect() {
        try {
            await this.producer.connect();
            await this.consumer.connect();
            this.admin = this.kafka.admin();
            await this.admin.connect();
            this.isConnected = true;
        } catch (error) {
            await new Promise(resolve => setTimeout(resolve, 5000));
            return await this.connect();
        }
    }

    async createTopics(topics) {
        if (!this.isConnected) return;
        try {
            const existingTopics = await this.admin.listTopics();
            const topicsToCreate = topics
                .filter(topic => !existingTopics.includes(topic))
                .map(topic => ({ topic, numPartitions: 1, replicationFactor: 1 }));

            if (topicsToCreate.length > 0) {
                await this.admin.createTopics({ topics: topicsToCreate, waitForLeaders: true });
            }
        } catch (error) {
            throw error;
        }
    }

    async publish(topic, message, headers = {}) {
        if (!this.isConnected) return;
        try {
            await this.producer.send({
                topic,
                messages: [{ value: JSON.stringify(message), headers }],
            });
        } catch (error) {}
    }

    async subscribe(topic, callback) {
        if (!this.isConnected) return;
        this.callbacks[topic] = callback;
        await this.consumer.subscribe({ topic, fromBeginning: true });
    }

    async startConsumer() {
        if (!this.isConnected) return;
        this.consumer.on(this.consumer.events.CRASH, () => process.exit(1));
        await this.consumer.run({
            eachMessage: async ({ topic, message }) => {
                const callback = this.callbacks[topic];
                if (callback) {
                    try {
                        const payload = JSON.parse(message.value.toString());
                        await callback(payload);
                    } catch (e) {}
                }
            },
        });
    }

    async disconnect() {
        try {
            await this.consumer.disconnect();
            await this.producer.disconnect();
            if (this.admin) await this.admin.disconnect();
        } catch (error) {}
    }
}

module.exports = new KafkaService();
