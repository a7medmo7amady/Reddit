const { Kafka } = require('kafkajs');

class KafkaService {
    constructor() {
        this.kafka = new Kafka({
            clientId: 'upload-service',
            brokers: (process.env.KAFKA_BROKERS || 'localhost:9092').split(','),
            retry: { initialRetryTime: 5000, retries: 20 },
            connectionTimeout: 60000,
            requestTimeout: 60000,
        });
        this.producer = this.kafka.producer();
        this.consumer = this.kafka.consumer({ 
            groupId: 'upload-processing-v3',
            sessionTimeout: 45000,
            rebalanceTimeout: 60000,
            heartbeatInterval: 3000
        });
        this.isConnected = false;
        this.callbacks = {};
    }

    async connect(maxRetries = 10) {
        let retries = 0;
        while (retries < maxRetries) {
            try {
                console.log(`[KafkaService] Connecting... (Attempt ${retries + 1}/${maxRetries})`);
                await this.producer.connect();
                await this.consumer.connect();
                this.admin = this.kafka.admin();
                await this.admin.connect();
                this.isConnected = true;
                console.log('[KafkaService] Connected successfully.');
                return;
            } catch (error) {
                retries++;
                console.error(`[KafkaService] Connection failed: ${error.message}. Retrying in 5s...`);
                await new Promise(resolve => setTimeout(resolve, 5000));
            }
        }
        console.error('[KafkaService] Max retries reached. Service will continue without Kafka features.');
    }

    async createTopics(topics, maxRetries = 5) {
        if (!this.isConnected) return;
        let retries = 0;
        while (retries < maxRetries) {
            try {
                const existingTopics = await this.admin.listTopics();
                const topicsToCreate = topics
                    .filter(topic => !existingTopics.includes(topic))
                    .map(topic => ({ topic, numPartitions: 1, replicationFactor: 1 }));

                if (topicsToCreate.length > 0) {
                    await this.admin.createTopics({ topics: topicsToCreate, waitForLeaders: true });
                }
                console.log('[KafkaService] Topics verified/created.');
                return;
            } catch (error) {
                retries++;
                console.warn(`[KafkaService] Topic creation attempt ${retries} failed: ${error.message}. Retrying...`);
                await new Promise(resolve => setTimeout(resolve, 5000));
            }
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
        this.consumer.on(this.consumer.events.CRASH, (e) => {
            console.error('[KafkaService] Consumer crashed:', e.error || e.message || e);
        });
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
