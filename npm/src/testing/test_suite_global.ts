import {TestSuite} from "./test_suite"
import type {TestFn} from "./test_utils"
import {report, ReportFormat} from "./test_reporter"

let GLOBAL_SUITE: TestSuite | undefined
let RUNNING_GLOBAL_SUITE = false

/**
 * Registers a test with the global test suite. When using this function, your
 * tests will automatically run on the next tick of the event loop. To group
 * tests, or control how and when they are run, use TestSuite.new().
 */
export function test(name: string, fn: TestFn) {
    if (!GLOBAL_SUITE) {
        initGlobalTestSuite()
        if (!process.argv.includes("--global")) {
            console.warn(
                `------------------------------------------------------------------
Warning: you called test() outside of a test suite. Your test will
be added to the global test suite.

Offending test: "%s"

To remove this warning, either:
    - Use TestSuite.new(cb)
    - Pass --global to ignore this warning
------------------------------------------------------------------`,
                name,
            )
        }
    }
    if (RUNNING_GLOBAL_SUITE) {
        console.error(
            `Refusing to register test: %s

This test was registered after the global test suite started running. This is
likely because this test was registered asynchronously.`,
            name,
        )
        process.exit(1)
    }

    GLOBAL_SUITE!.test(name, fn)
}

/**
 * Makes the bare `test` import work by creating a global test suite. This is
 * an alternative to defining an explicit test suite:
 *
 * @example
 * suite(test => {
 *   test("a", () => { ... })
 *   test("b", () => { ... })
 *   test("c", () => { ... })
 * })
 *
 * // becomes
 * initGlobalTestSuite()
 * test("a", () => { ... })
 * test("b", () => { ... })
 * test("c", () => { ... })
 */
export function initGlobalTestSuite() {
    GLOBAL_SUITE = new TestSuite()

    // We don't know precisely when the last test will be defined, so just assume
    // that they'll all be ready by the next tick of the event loop.
    Promise.resolve().then(async () => {
        RUNNING_GLOBAL_SUITE = true
        const results = await GLOBAL_SUITE!.run()

        let format: ReportFormat
        if (process.argv.includes("--format=json")) {
            format = ReportFormat.JSON
        } else {
            format = ReportFormat.Pretty
        }

        report(results, {format})
    })
}
