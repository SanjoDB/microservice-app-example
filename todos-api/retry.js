'use strict';

class RetryConfig {
    constructor(maxAttempts = 3, initialWaitTime = 100, maxWaitTime = 2000) {
        this.maxAttempts = maxAttempts;
        this.waitTime = initialWaitTime;
        this.maxWaitTime = maxWaitTime;
    }
}

async function retry(config, operation) {
    let waitTime = config.waitTime;
    let lastError;

    for (let attempt = 1; attempt <= config.maxAttempts; attempt++) {
        try {
            return await operation();
        } catch (error) {
            lastError = error;
            console.log(`Retry attempt ${attempt}/${config.maxAttempts} failed: ${error.message}`);

            if (attempt === config.maxAttempts) {
                break;
            }

            // Exponential backoff
            if (waitTime < config.maxWaitTime) {
                waitTime *= 2;
            }

            await new Promise(resolve => setTimeout(resolve, waitTime));
        }
    }

    throw lastError;
}

module.exports = {
    RetryConfig,
    retry
};