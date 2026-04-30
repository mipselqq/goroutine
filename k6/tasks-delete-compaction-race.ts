import { check } from 'k6';
import http from 'k6/http';
import { defaultRegisterAndLogin, createBoard, createColumn, createTaskRequest, deleteBoard, deleteTaskRequest, getTask } from './prelude.ts';
import type { TasksDeleteCompactionSetup } from './types.ts';

const NUM_TASKS = 20 + 1;

export function setup(): TasksDeleteCompactionSetup {
    const authHeader = defaultRegisterAndLogin();
    const boardId = createBoard(authHeader);
    const columnId = createColumn(boardId, authHeader);
    const testTaskIds = http.batch(
        Array.from({ length: NUM_TASKS }, () => createTaskRequest(boardId, columnId, authHeader)),
    ).map((response) => (
        check(response, { 'create task batch is all 201': (x) => x.status === 201 }),
        response.json('id') as string
    ));

    const lastTaskId = testTaskIds.pop() as string;

    return { authHeader, boardId, columnId, testTaskIds, lastTaskId };
}

export const options = {
    scenarios: {
        tasksDeleteCompactionRace: {
            executor: 'per-vu-iterations',
            vus: 1,
            iterations: 1,
        },
    },
    thresholds: {
        checks: ['rate == 1'],
    },
}

export default function tasksDeleteCompactionRace({ authHeader, boardId, columnId, testTaskIds }: TasksDeleteCompactionSetup): void {
    const batch = http.batch(
        testTaskIds.map((taskId) => deleteTaskRequest(boardId, columnId, taskId, authHeader)),
    );

    for (const response of batch) {
        check(response, { '204': (x) => x.status === 204 });
    }
}

export function teardown({ authHeader, boardId, columnId, lastTaskId }: TasksDeleteCompactionSetup): void {
    const lastTask = getTask(boardId, columnId, lastTaskId, authHeader);
    check(lastTask, { 'left task position is 1': (x) => x.position === 1 });
    deleteBoard(boardId, authHeader);
}
