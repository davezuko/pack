# pack

## Testing

```ts
import {test} from "@TODO"

test("simple test", (t) => {
    const a = 1
    const b = 2
    if (a !== b) {
        t.error("expected %s to === %s", a, b)
    }
})

// Asynchronous tests will wait for the promise to settle before finishing.
// If the promise rejects, the test is marked as failed.
test("simple test", (t) => {
    const a = 1
    const b = 2
    if (a !== b) {
        t.error("expected %s to === %s", a, b)
    }
})
```

The core of the testing library is built around a simple interface.

```ts
{
    /** Records a message in the test's output log. */
    log(msg: string, ...args: any[]): void

    /** Equivalent to t.log() followed by t.fail(). */
    error(msg: string, ...args: any[]): void

    /** Equivalent to t.log() followed by t.failNow(). */
    fatal(msg: string, ...args: any[]): void

    /** Skips the current test. */
    skip(): void

    /** Marks the test as failed but continues execution. */
    fail(): void

    /** Marks the test as failed and stops execution. */
    failNow(): void
}
```

Assertion helpers are available as a separate module. These are built on top
of the core testing library interface. Feel free to create your own utilities
as well!

```ts
import {test, assert} from "@TODO"

test("strict equality (===)", (t) => {
    const a = {}
    const b = a
    assert.is(t, a, b) // pass
    assert.is(t, a, {}) // fail
})

test("deep equality", (t) => {
    const a = {}
    const b = {}
    assert.eq(t, a, b) // pass
    assert.eq(t, a, {foo: "bar"}) // fail
})

// You can bind assert to `t` to make writing assertions more convenient:
test("bound asset", (t) => {
    const assert = assert.bind(t)
    assert.eq(a, b) // notice how we omit the first argument
})
```
