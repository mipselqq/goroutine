import http, { RefinedResponse } from 'k6/http';
import type { AuthHeader, K6Response } from './types.ts';

export const API_BASE = __ENV.K6_ROOT || 'http://localhost:8080';
export const PWD = 'testPassword$123';
export const JSON_HEADER = { 'Content-Type': 'application/json' };

export function generateTimedEmail(): string {
    return `${Date.now()}@t.t`;
}

export function createBoard(authHeader: AuthHeader): string {
    const boardResp = http.post(
        `${API_BASE}/v1/boards`,
        JSON.stringify({ name: 'Test board', description: '' }),
        { headers: authHeader },
    );
    if (boardResp.status !== 201) {
        throw new Error(`create board failed: ${boardResp.status} ${boardResp.body}`);
    }

    return boardResp.json('id') as string;
}

export function createColumnRequest(boardId: string, authHeader: AuthHeader) {
    return {
        method: 'POST',
        url: `${API_BASE}/v1/boards/${boardId}/columns`,
        params: { headers: { ...JSON_HEADER, ...authHeader } },
        body: JSON.stringify({ name: 'Test column', description: '' }),
    };
}

export function createColumn(boardId: string, authHeader: AuthHeader): string {
    const req = createColumnRequest(boardId, authHeader);

    const columnResp = http.request(req.method, req.url, req.body, req.params);
    if (columnResp.status !== 201) {
        throw new Error(`create column failed: ${columnResp.status} ${columnResp.body}`);
    }

    return columnResp.json('id') as string;
}

export function deleteBoard(boardId: string, authHeader: AuthHeader): void {
    const deleteResp = http.del(
        `${API_BASE}/v1/boards/${boardId}`,
        null,
        { headers: authHeader },
    );
    if (deleteResp.status !== 204) {
        throw new Error(`delete board failed: ${deleteResp.status} ${deleteResp.body}`);
    }
}

export function createTaskRequest(boardId: string, columnId: string, authHeader: AuthHeader) {
    return {
        method: 'POST',
        url: `${API_BASE}/v1/boards/${boardId}/columns/${columnId}/tasks`,
        params: { headers: { ...JSON_HEADER, ...authHeader } },
        body: JSON.stringify({ name: 'Test task', description: '' }),
    };
}
export function createTask(boardId: string, columnId: string, authHeader: AuthHeader): string {
    const req = createTaskRequest(boardId, columnId, authHeader);

    const taskResp = http.request(req.method, req.url, req.body, req.params);
    if (taskResp.status !== 201) {
        throw new Error(`create task failed: ${taskResp.status} ${taskResp.body}`);
    }
    return taskResp.json('id') as string;
}

export function defaultRegisterAndLogin(): AuthHeader {
    const email = generateTimedEmail();
    const password = PWD;

    const registerResp = http.post(
        `${API_BASE}/v1/register`,
        JSON.stringify({ email, password }),
        { headers: { 'Content-Type': 'application/json' } },
    );
    if (registerResp.status !== 200) {
        throw new Error(`register failed: ${registerResp.status} ${registerResp.body}`);
    }

    const loginResp = http.post(
        `${API_BASE}/v1/login`,
        JSON.stringify({ email, password }),
        { headers: { 'Content-Type': 'application/json' } },
    );
    if (loginResp.status !== 200) {
        throw new Error(`login failed: ${loginResp.status} ${loginResp.body}`);
    }

    const token = loginResp.json('token');

    return { Authorization: `Bearer ${token}` };
}
