import {TestResult, TestResultStatus} from "./test_utils"

export enum ReportFormat {
    JSON = "json",
    Pretty = "pretty",
}

export function report(
    results: TestResult[],
    opts: {
        format: ReportFormat
    },
) {
    switch (opts.format) {
        case ReportFormat.JSON: {
            const json = results.map((result) => ({
                name: result.name,
                stat: result.stat,
                logs: result.logs,
            }))
            console.log(JSON.stringify(json))
            break
        }
        case ReportFormat.Pretty: {
            for (const result of results) {
                switch (result.stat) {
                    case TestResultStatus.Pass:
                        console.log("[pass] %s", result.name)
                        break
                    case TestResultStatus.Skip:
                        console.log("[skip] %s", result.name)
                        break
                    case TestResultStatus.Fail:
                        console.log("[fail] %s", result.name)
                        for (const log of result.logs) {
                          console.log("  %s", log)
                        }
                        break
                }
            }
            break
        }
    }
}
