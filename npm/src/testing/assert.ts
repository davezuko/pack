// @ts-expect-error
import * as eq from "lodash.isequal"
import type {TestUtils} from "./test_utils"

// NOTE: all static methods are explicitly defined as instance methods as well
// so that TypeScript provides proper documentation on an instance. If there is
// a better way to accomplish this I'm all ears!
export class Assert {
    private t: TestUtils

    constructor(t: TestUtils) {
        this.t = t
    }

    /**
     * Creates a new instance of Assert bound to a specific test instance of
     * TestUtils. This is useful if you don't want to supply a TestUtils
     * instance as the first argument of every assertion.
     *
     * @example
     * Assert.is(t, 1, 1)
     * Assert.is(t, 2, 2)
     * Assert.is(t, 3, 3)
     *
     * // becomes
     * const assert = assert.bind(t)
     * assert.is(1, 1)
     * assert.is(2, 2)
     * assert.is(3, 3)
     */
    static bind(t: TestUtils) {
        return new Assert(t)
    }

    /**
     * Asserts that a is deeply equal to b.
     */
    eq<T>(a: T, b: T) {
        Assert.eq(this.t, a, b)
    }

    /**
     * Asserts that key exists in obj.
     */
    has(obj: any, key: string) {
        Assert.has(this.t, obj, key)
    }

    /**
     * Asserts that x exists in coll.
     */
    includes<T>(coll: T[], x: T) {
        Assert.includes(this.t, coll, x)
    }

    /**
     * Asserts that x exists in coll.
     */
    includesEq<T>(coll: T[], x: T) {
        Assert.includesEq(this.t, coll, x)
    }

    /**
     * Asserts that a is strictly equal to b (===).
     */
    is<T>(a: T, b: T) {
        Assert.is(this.t, a, b)
    }

    /**
     * Asserts that a is deeply equal to b.
     */
    static eq<T>(t: TestUtils, a: T, b: T) {
        if (!eq(a, b)) {
            t.error("expected %s to be deeply equal to %s", a, b)
        }
    }

    /**
     * Asserts that key exists in obj.
     */
    static has(t: TestUtils, obj: any, key: string) {
        if (!Object.hasOwnProperty.call(obj, key)) {
            t.error("expected %s to have key %s", obj, key)
        }
    }

    /**
     * Asserts that x exists in coll.
     */
    static includes<T>(t: TestUtils, coll: T[], x: T) {
        if (!coll.includes(x)) {
            t.error("expected %s to contain %s", coll, x)
        }
    }

    /**
     * Asserts that x exists in coll.
     */
    static includesEq<T>(t: TestUtils, coll: T[], x: T) {
        let found = false
        for (const item of coll) {
            if (eq(item, x)) {
                found = true
                break
            }
        }
        if (!found) {
            t.error("expected %s to contain %s", coll, x)
        }
    }

    /**
     * Asserts that a is strictly equal to b (===).
     */
    static is<T>(t: TestUtils, a: T, b: T) {
        if (a !== b) {
            t.error("expected %s to be strictly equal to %s", a, b)
        }
    }
}