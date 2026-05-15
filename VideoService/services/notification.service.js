/**
 * notification.service.js
 *
 * Sends notifications to the notification-service via HTTP.
 */

const consulService = require('./consul.service');

const NOTIFICATION_SERVICE_NAME = process.env.NOTIFICATION_SERVICE_NAME || 'notification';

async function sendNotification({ userId, title, message, link, type }) {
    try {
        let url = await consulService.resolve(NOTIFICATION_SERVICE_NAME);
        if (!url) {
            url = process.env.NOTIFICATION_SERVICE_URL || 'http://notification-service:8084';
        }

        const response = await fetch(`${url}/api/v1/notifications/send`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                user_id: userId,
                title,
                message,
                link: link || '',
                type: type || 'REPLY'
            })
        });

        if (!response.ok) {
            const text = await response.text();
            console.warn(`[Notification] Failed to send notification: ${response.status} ${text}`);
        } else {
            console.log(`[Notification] Sent to user ${userId}: ${title}`);
        }
    } catch (err) {
        console.warn(`[Notification] Error sending notification: ${err.message}`);
    }
}

module.exports = { sendNotification };
