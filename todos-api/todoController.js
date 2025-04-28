'use strict';
const cache = require('memory-cache');
const {Annotation, jsonEncoder: {JSON_V2}} = require('zipkin');
const { RetryConfig, retry } = require('./retry');
const CircuitBreaker = require('./circuit_breaker');

const OPERATION_CREATE = 'CREATE',
      OPERATION_DELETE = 'DELETE';

class TodoController {
    constructor({tracer, redisClient, logChannel}) {
        this._tracer = tracer;
        this._redisClient = redisClient;
        this._logChannel = logChannel;
        this._retryConfig = new RetryConfig(3, 100, 2000);
        
        // Inicializar los circuit breakers para cada operación
        this._redisCircuitBreaker = new CircuitBreaker(3, 10000);
        this._cacheCircuitBreaker = new CircuitBreaker(3, 10000);
    }

    async _logOperation(opName, username, todoId) {
        return this._redisCircuitBreaker.execute(async () => {
            return retry(this._retryConfig, () => {
                return new Promise((resolve, reject) => {
                    this._tracer.scoped(() => {
                        const traceId = this._tracer.id;
                        this._redisClient.publish(this._logChannel, JSON.stringify({
                            zipkinSpan: traceId,
                            opName: opName,
                            username: username,
                            todoId: todoId,
                        }), (err) => {
                            if (err) reject(err);
                            else resolve();
                        });
                    });
                });
            });
        });
    }

    async _getTodoData(userID) {
        return this._cacheCircuitBreaker.execute(async () => {
            return retry(this._retryConfig, () => {
                return new Promise((resolve) => {
                    const data = cache.get(userID);
                    if (data == null) {
                        const newData = {
                            items: {
                                '1': {
                                    id: 1,
                                    content: "Create new todo",
                                },
                                '2': {
                                    id: 2,
                                    content: "Update me",
                                },
                                '3': {
                                    id: 3,
                                    content: "Delete example ones",
                                }
                            },
                            lastInsertedID: 3
                        };
                        this._setTodoData(userID, newData);
                        resolve(newData);
                    } else {
                        resolve(data);
                    }
                });
            });
        });
    }

    async _setTodoData(userID, data) {
        return this._cacheCircuitBreaker.execute(async () => {
            return retry(this._retryConfig, () => {
                return new Promise((resolve) => {
                    cache.put(userID, data);
                    resolve();
                });
            });
        });
    }

    async list(req, res) {
        try {
            const data = await this._getTodoData(req.user.username);
            res.json(data.items);
        } catch (error) {
            console.error('Error in list operation:', error);
            res.status(500).json({ error: 'Internal server error' });
        }
    }

    async create(req, res) {
        try {
            const data = await this._getTodoData(req.user.username);
            const todo = {
                content: req.body.content,
                id: data.lastInsertedID + 1
            };
            data.items[data.lastInsertedID + 1] = todo;
            data.lastInsertedID++;
            
            await this._setTodoData(req.user.username, data);
            await this._logOperation(OPERATION_CREATE, req.user.username, todo.id);
            
            res.json(todo);
        } catch (error) {
            console.error('Error in create operation:', error);
            res.status(500).json({ error: 'Internal server error' });
        }
    }

    async delete(req, res) {
        try {
            const data = await this._getTodoData(req.user.username);
            const id = req.params.taskId;
            delete data.items[id];
            
            await this._setTodoData(req.user.username, data);
            await this._logOperation(OPERATION_DELETE, req.user.username, id);
            
            res.status(204).send();
        } catch (error) {
            console.error('Error in delete operation:', error);
            res.status(500).json({ error: 'Internal server error' });
        }
    }
}

module.exports = TodoController;