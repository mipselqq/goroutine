import { Trend, Rate } from 'k6/metrics';
import { sleep } from 'k6';
import {
    defaultRegisterAndLogin, createBoard, createColumn, createTask,
    getAggregate, getBoard, listTasks,
    moveTask, deleteColumn, deleteTask, updateTask, deleteBoard,
} from './prelude.ts';
import type { AuthHeader } from './types.ts';

const L1_WAIT = 0.5;
const L2_WAIT = 1;
const L3_WAIT = 2;
const L4_WAIT = 4;

const VUS_STEP = parseInt(__ENV.K6_VUS_STEP || '500');
const VUS_PLATEU_DURATION = __ENV.K6_VUS_PLATEU_DURATION || '60s';
const RAMP_DURATION = __ENV.K6_RAMP_DURATION || '5s';
const AFTER_FAIL_DURATION = __ENV.K6_AFTER_FAIL_DURATION || '30s';
const DELAY_ABORT_EVAL = __ENV.K6_DELAY_ABORT_EVAL || '30s';
const MAX_STAGES = parseInt(__ENV.K6_MAX_STAGES || '50');

const m = {
    createBoard: { d: new Trend('create_board_duration'), e: new Rate('create_board_errors') },
    createColumn: { d: new Trend('create_column_duration'), e: new Rate('create_column_errors') },
    createTask: { d: new Trend('create_task_duration'), e: new Rate('create_task_errors') },
    aggregate: { d: new Trend('aggregate_duration'), e: new Rate('aggregate_errors') },
    getBoard: { d: new Trend('get_board_duration'), e: new Rate('get_board_errors') },
    listTasks: { d: new Trend('list_tasks_duration'), e: new Rate('list_tasks_errors') },
    moveTask: { d: new Trend('move_task_duration'), e: new Rate('move_task_errors') },
    deleteColumn: { d: new Trend('delete_column_duration'), e: new Rate('delete_column_errors') },
    deleteTask: { d: new Trend('delete_task_duration'), e: new Rate('delete_task_errors') },
    updateTask: { d: new Trend('update_task_duration'), e: new Rate('update_task_errors') },
    deleteBoard: { d: new Trend('delete_board_duration'), e: new Rate('delete_board_errors') },
};

function measure(metric: { d: Trend; e: Rate }, fn: () => void): void {
    const start = Date.now();
    try {
        fn();
        metric.e.add(0);
    } catch (error) {
        metric.e.add(1);
        throw new Error('measure() fn threw an error', { cause: error });
    }
    metric.d.add(Date.now() - start);
}

export const options = {
    // The user gets angry here
    thresholds: {
        http_req_failed: [
            { threshold: 'rate < 0.01', abortOnFail: true, delayAbortEval: DELAY_ABORT_EVAL },
        ],
        http_req_duration: [
            { threshold: 'p(95) < 1000', abortOnFail: true, delayAbortEval: DELAY_ABORT_EVAL },
        ],
    },
    scenarios: {
        rampToBreak: {
            executor: 'ramping-vus',
            startVUs: 0,
            gracefulStop: AFTER_FAIL_DURATION,
            stages: Array.from( // Ramp until the server gives up
                { length: MAX_STAGES },
                (_, i) => [
                    { duration: RAMP_DURATION, target: VUS_STEP * (i + 1) },
                    { duration: VUS_PLATEU_DURATION, target: VUS_STEP * (i + 1) },
                ]
            ).flat(),
        },
    },
};

// One-time registration for ALL VUs as we don't test the slow register here
export function setup(): { auth: AuthHeader } {
    const auth = defaultRegisterAndLogin();
    return { auth };
}

let boardIds: string[] = [];
let colIds: string[][] = [];

export default function realisticHappyPath({ auth }: { auth: AuthHeader }): void {
    // Creates 4 boards, 5 columns per board, 10 tasks per column
    sleep(L4_WAIT);
    boardIds = [];
    colIds = [];
    for (let b = 0; b < 4; b++) {
        let bid = '';
        measure(m.createBoard, () => { bid = createBoard(auth); });
        boardIds.push(bid);
        sleep(L3_WAIT);

        colIds[b] = [];
        for (let c = 0; c < 5; c++) {
            let cid = '';
            measure(m.createColumn, () => { cid = createColumn(bid, auth); });
            colIds[b].push(cid);
            sleep(L3_WAIT);

            for (let t = 0; t < 10; t++) {
                measure(m.createTask, () => { createTask(bid, cid, auth); });
                sleep(L3_WAIT);
            }
        }
    }

    // Looks at each board through aggregate method, kinda refreshing page
    for (let r = 0; r < 3; r++) {
        for (const bid of boardIds) {
            measure(m.aggregate, () => { getAggregate(bid, auth); });
            sleep(L2_WAIT);
        }
    }

    // Looks at first board
    measure(m.getBoard, () => { getBoard(boardIds[0], auth); });
    sleep(L2_WAIT);

    // Moves all tasks from 1st to 5th column
    const col1 = colIds[0][0];
    const col5 = colIds[0][4];
    let col1Tasks: { id: string }[] = [];
    measure(m.listTasks, () => { col1Tasks = listTasks(boardIds[0], col1, auth); });
    for (let i = 0; i < col1Tasks.length; i++) {
        const pos = i + 1;
        measure(m.moveTask, () => { moveTask(boardIds[0], col1, col1Tasks[i].id, col5, pos, auth); });
        sleep(L1_WAIT);
    }

    // Deletes 2nd column
    measure(m.deleteColumn, () => { deleteColumn(boardIds[0], colIds[0][1], auth); });
    sleep(L1_WAIT);

    // Removes each task from 5th column
    let col5Tasks: { id: string }[] = [];
    measure(m.listTasks, () => { col5Tasks = listTasks(boardIds[0], col5, auth); });
    for (const t of col5Tasks) {
        measure(m.deleteTask, () => { deleteTask(boardIds[0], col5, t.id, auth); });
        sleep(L1_WAIT);
    }

    // Renames each task in 3rd column
    const col3 = colIds[0][2];
    let col3Tasks: { id: string }[] = [];
    measure(m.listTasks, () => { col3Tasks = listTasks(boardIds[0], col3, auth); });
    for (const t of col3Tasks) {
        measure(m.updateTask, () => { updateTask(boardIds[0], col3, t.id, 'Renamed task', '', auth); });
        sleep(L1_WAIT);
    }

    // Deletes all boards sequentially and repeats the whole process until the end of the test
    for (const bid of boardIds) {
        measure(m.deleteBoard, () => { deleteBoard(bid, auth); });
    }
}
