import { check } from 'k6';
import http from 'k6/http';
import { defaultRegisterAndLogin, createBoard, createColumnRequest, deleteBoard, deleteColumnRequest, getColumn } from './prelude.ts';
import type { ColumnsDeleteCompactionSetup } from './types.ts';

const NUM_COLUMNS = 20 + 1;

export function setup(): ColumnsDeleteCompactionSetup {
    const authHeader = defaultRegisterAndLogin();
    const boardId = createBoard(authHeader);
    const testColumnIds = http.batch(
        Array.from({ length: NUM_COLUMNS }, () => createColumnRequest(boardId, authHeader)),
    ).map((response) => (
        check(response, { 'create column batch is all 201': (x) => x.status === 201 }),
        response.json('id') as string
    ));

    const lastColumnId = testColumnIds.pop() as string;

    return { authHeader, boardId, testColumnIds, lastColumnId };
}

export const options = {
    scenarios: {
        columnsDeleteCompactionRace: {
            executor: 'per-vu-iterations',
            vus: 1,
            iterations: 1,
        },
    },
    thresholds: {
        checks: ['rate == 1'],
    },
}

export default function columnsDeleteCompactionRace({ authHeader, boardId, testColumnIds }: ColumnsDeleteCompactionSetup): void {
    const batch = http.batch(
        testColumnIds.map((columnId) => deleteColumnRequest(boardId, columnId, authHeader)),
    );

    for (const response of batch) {
        check(response, { '204': (x) => x.status === 204 });
    }
}

export function teardown({ authHeader, boardId, lastColumnId }: ColumnsDeleteCompactionSetup): void {
    const lastColumn = getColumn(boardId, lastColumnId, authHeader);
    check(lastColumn, { 'left column position is 1': (x) => x.position === 1 });
    deleteBoard(boardId, authHeader);
}
