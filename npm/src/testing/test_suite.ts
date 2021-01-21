import {Test, TestUtils, TestResult, TestFn} from "./test_utils"

export class TestSuite {
    private tests: Test[]

    constructor() {
        this.tests = []
    }

    /**
     * Creates and runs a new test suite. You should define tests within the
     * callback function.
     *
     * @example
     * TestSuite.new((test) => {
     *   test("my-test", t => {
     *     t.fail()
     *   })
     * })
     */
    static new(
        fn: (test: (name: string, fn: TestFn) => void) => void,
    ): TestSuite {
        const suite = new TestSuite()
        fn(suite.test.bind(suite))
        return suite
    }

    /**
     * Registers a new test with the suite.
     */
    test(name: string, fn: TestFn) {
        this.tests.push(new Test(name, fn))
    }

    /**
     * Runs all tests registered in the suite.
     */
    run(): Promise<TestResult[]> {
        if (!this.tests.length) {
            throw new Error("No tests to run")
        }

        return Promise.all(
            this.tests.map((test) => {
                const t = new TestUtils(test.name)
                return test.run(t)
            }),
        )
    }
}
