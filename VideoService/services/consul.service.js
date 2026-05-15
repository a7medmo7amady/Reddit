/**
 * consul.service.js
 *
 * Lightweight Consul HTTP API wrapper for Node.js services.
 * - Registers this service with Consul on startup
 * - Resolves other services by name with healthy-instance selection
 * - Falls back to static env var addresses if Consul is unreachable
 */

const http = require('http');

const CONSUL_ADDR = process.env.CONSUL_ADDR || 'consul:8500';
const SERVICE_NAME = process.env.SERVICE_NAME || 'video';
const SERVICE_PORT = parseInt(process.env.PORT || '8083', 10);
const SERVICE_HOST = process.env.HOSTNAME || 'video-service'; // Docker sets HOSTNAME to container name

// ── Consul HTTP helper ────────────────────────────────────────────────────────

function consulRequest(method, path, body) {
    return new Promise((resolve, reject) => {
        const [host, port] = CONSUL_ADDR.split(':');
        const payload = body ? JSON.stringify(body) : null;

        const options = {
            hostname: host,
            port: port || 8500,
            path,
            method,
            headers: {
                'Content-Type': 'application/json',
                ...(payload ? { 'Content-Length': Buffer.byteLength(payload) } : {}),
            },
        };

        const req = http.request(options, (res) => {
            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => {
                try { resolve({ status: res.statusCode, body: data ? JSON.parse(data) : null }); }
                catch { resolve({ status: res.statusCode, body: data }); }
            });
        });

        req.on('error', reject);
        if (payload) req.write(payload);
        req.end();
    });
}

// ── Registration ──────────────────────────────────────────────────────────────

async function register() {
    try {
        const result = await consulRequest('PUT', '/v1/agent/service/register', {
            ID:      `${SERVICE_NAME}-${SERVICE_PORT}`,
            Name:    SERVICE_NAME,
            Address: SERVICE_HOST,
            Port:    SERVICE_PORT,
            Tags:    ['http'],
            Check: {
                HTTP:     `http://${SERVICE_HOST}:${SERVICE_PORT}/health`,
                Interval: '10s',
                Timeout:  '5s',
                DeregisterCriticalServiceAfter: '30s',
            },
        });

        if (result.status === 200) {
            console.log(`[Consul] Registered ${SERVICE_NAME} (${SERVICE_HOST}:${SERVICE_PORT})`);
        } else {
            console.warn(`[Consul] Registration returned ${result.status}`);
        }
    } catch (err) {
        console.warn(`[Consul] Registration failed (${err.message}) — running without service discovery`);
    }
}

async function deregister() {
    try {
        await consulRequest('PUT', `/v1/agent/service/deregister/${SERVICE_NAME}-${SERVICE_PORT}`);
        console.log('[Consul] Deregistered');
    } catch (err) {
        console.warn('[Consul] Deregister failed:', err.message);
    }
}

// ── Resolution ────────────────────────────────────────────────────────────────

/** Static fallback map — add known services here */
const STATIC_FALLBACKS = {
    user:         `http://${process.env.USER_SERVICE_ADDR  || 'user-service:8080'}`,
    feed:         `http://${process.env.FEED_SERVICE_ADDR  || 'feed-service:8081'}`,
    notification: `http://${process.env.NOTIFICATION_SERVICE_ADDR || 'notification-service:8084'}`,
    'api-gateway': `http://${process.env.API_GATEWAY_URL?.replace(/^https?:\/\//, '') || 'api-gateway:8088'}`,
};

/**
 * Returns a healthy instance URL ("http://host:port") for the given service name.
 * Falls back to the static map if Consul is unavailable.
 */
async function resolve(serviceName) {
    try {
        const result = await consulRequest('GET',
            `/v1/health/service/${encodeURIComponent(serviceName)}?passing=true`);

        if (result.status === 200 && Array.isArray(result.body) && result.body.length > 0) {
            const entry = result.body[Math.floor(Math.random() * result.body.length)];
            const host = entry.Service.Address || entry.Node.Address;
            const port = entry.Service.Port;
            const url = `http://${host}:${port}`;
            console.log(`[Consul] resolved ${serviceName} → ${url}`);
            return url;
        }
    } catch (err) {
        console.warn(`[Consul] lookup "${serviceName}" failed (${err.message}), using fallback`);
    }

    const fallback = STATIC_FALLBACKS[serviceName];
    if (fallback) {
        return fallback;
    }
    console.warn(`[Consul] no fallback defined for "${serviceName}"`);
    return null;
}

module.exports = { register, deregister, resolve };
