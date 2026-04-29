import { check } from 'k6';
import http from 'k6/http';
import { defaultRegisterAndLogin, createBoard, deleteBoard, API_BASE, JSON_HEADER } from './prelude.ts';
import type { ColumnsRaceSetup } from './types.ts';

export function setup(): ColumnsRaceSetup {
    const authHeader = defaultRegisterAndLogin();
    const boardId = createBoard(authHeader);
    return { authHeader, boardId };
}

export const options = {
    scenarios: {
        columnsRace: {
            // Very easy to reproduce the race condition
            // https://github.com/mipselqq/goroutine/pull/173
            executor: 'per-vu-iterations',
            vus: 1,
            iterations: 1,
            maxDuration: '1s',
        },
    },
    thresholds: {
        checks: ['rate == 1'],
    },
};

export default function columnsRace({ authHeader, boardId }: ColumnsRaceSetup): void {
    const batch = http.batch(Array.from({ length: 10 }, () => ({
        method: 'POST',
        url: `${API_BASE}/v1/boards/${boardId}/columns`,
        body: JSON.stringify({ name: `t-${Date.now()}`, description: '' }),
        params: {
            headers: {
                ...authHeader,
                ...JSON_HEADER,
            },
        },
    })));

    for (const response of batch) {
        check(response, { '201': (x) => x.status === 201 });
    }
}

export function teardown({ authHeader, boardId }: ColumnsRaceSetup): void {
    deleteBoard(boardId, authHeader);
}
