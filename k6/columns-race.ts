import { check } from 'k6';
import http from 'k6/http';
import { defaultRegisterAndLogin, createBoard, createColumnRequest, deleteBoard } from './prelude.ts';
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

// Attempting go catch 500 Internal Error because of Unique Violation
export default function columnsRace({ authHeader, boardId }: ColumnsRaceSetup): void {
    const batch = http.batch(
        Array.from({ length: 10 }, () => createColumnRequest(boardId, authHeader)),
    );

    for (const response of batch) {
        check(response, { '201': (x) => x.status === 201 });
    }
}

export function teardown({ authHeader, boardId }: ColumnsRaceSetup): void {
    deleteBoard(boardId, authHeader);
}
