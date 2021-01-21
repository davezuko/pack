import * as util from "util"

// control flow in tests is managed with exceptions, since user code
// is not rewritten (yet). These objects act as markers so that we
// can discern why a test failed.
const TEST_FATAL = {}
const TEST_SKIP = {}
const TEST_UNCAUGHT_EXCEPTION = {}

export type TestFn = (t: TestUtils) => void

export interface TestResult {
    name: string
    stat: TestResultStatus
    logs: string[]
}

export enum TestResultStatus {
    Pass = "pass",
    Fail = "fail",
    Skip = "skip",
}

export class Test {
    name: string
    private fn: TestFn

    constructor(name: string, fn: TestFn) {
        this.name = name
        this.fn = fn
    }

    async run(t: TestUtils): Promise<TestResult> {
        let stat = TestResultStatus.Pass
        try {
            await this.fn(t)
            t.done = true
        } catch (e) {
            switch (e) {
                case TEST_FATAL:
                    stat = TestResultStatus.Fail
                    break
                case TEST_SKIP:
                    stat = TestResultStatus.Skip
                    break
                default:
                    t.log(e)
                    t.failed = true
                    throw TEST_UNCAUGHT_EXCEPTION
            }
        }
        if (t.failed) {
            stat = TestResultStatus.Fail
        }

        return {
            name: this.name,
            stat: stat,
            logs: t.logs,
        }
    }
}

export class TestUtils {
    name: string
    logs: string[]
    failed: boolean
    done: boolean

    constructor(name: string) {
        this.name = name
        this.logs = []
        this.failed = false
        this.done = false
    }

    private assertNotDone(method: string) {
        if (this.done) {
            console.error(
                'Attempted to call %s from test("%s") after the test had finished running. This is likely because your test initiated an asynchronous operation but did not wait for it to complete.',
                method,
                this.name,
            )
            process.exit(1)
        }
    }

    /**
     * Records a message in the test's output log.
     */
    log(msg: string, ...args: any[]) {
        this.assertNotDone("log")
        this.logs.push(util.format(msg, ...args))
    }

    /**
     * Equivalent to t.log() followed by t.fail().
     */
    error(msg: string, ...args: any[]) {
        this.assertNotDone("error")
        this.log(msg, ...args)
        this.fail()
    }

    /**
     * Equivalent to t.log() followed by t.failNow().
     */
    fatal(msg: string, ...args: any[]) {
        this.assertNotDone("fatal")
        this.log(msg, ...args)
        this.failNow()
    }

    /**
     * Skips the current test.
     */
    skip() {
        this.assertNotDone("skip")
        throw TEST_SKIP
    }

    /**
     * Marks the test as failed but continues execution.
     */
    fail() {
        this.assertNotDone("fail")
        this.failed = true
    }

    /**
     * Marks the test as failed and stops execution.
     */
    failNow() {
        this.assertNotDone("failNow")
        this.failed = true
        throw TEST_FATAL
    }
}
