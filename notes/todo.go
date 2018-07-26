package notes

/**
 * NEXT:: SEE MVP
 *
 * Note: after MVP is done, should focus on creating a new project that uses this
 * framework to implement an existing Fuchsia recipe + tests.
 */

/**
 * DONE 1:: Attach stderr stdout to expectation step output
 */

/**
 * DONE 2:: Wrap invocations serializable JSON object so logging is better
 */

/**
 * 4:: Try porting a simple recipe or module from luci-py and get it working
 */

/**
 * DONE 5:: Implement output verification.
 * The framework should check that the outputs declared by a step actually exist after the
 * step is run.
 */

/**
 * OBSOLETE Only structs are allowed for now
 *
 * 6:: Ensure function (adapter) implementations of Step are mockable
 */

/**
 * MVP::7:: Implement path checks / warnings
 *
 * Attempting to use a path that hasn't been declared in the outputs of step that has already
 * run, the framework should issue a warning that the path may not exist.
 */

/**
 * DONE 8:: Add a step name for easy reading and logging.
 */

/**
 * DONE 9:: Rename StepProvider to Step
 */

/**
 * DONE::10:: Add auto-test that expectation JSON is valid
 */

/**
 * 11:: Document and Enforce that step implementations must be Struct-types, simply because
 * this makes mocking steps much easier:  When treating mock zero-values as "anything",
 * we overwrite all of the zero-value fields in the mock with their counterparts in
 * the step and then perform a deep comparison to see if the mock matches.  We'd need
 * to write a different algorithm to perform this comparison on a function mock, and
 * we're lazy.
 */

/**
 * DONE::12:: Reach 90% Test coverage
 *
 * Didn't reach 90, but coverage is good enough for now.  Should keep adding tests as I go
 */

/**
 * 13:: Figure out what else is missing before I can make a reasonable demo
 *
 * Some things that come to mind:
 * - LUCI auth
 * - Logging to logdog
 * - CIPD support
 * - Tar support
 * - Some equivalent of recipe "properties"
 */

/**
 * 14:: Create some benchmarks for an equivalent recipe ported to chow.
 * - Also take a look at bootstrap time (time to fetch engine)
 * - Take a look at local execution time (might not matter as much)
 * - Take a look at test execution time (might not matter as much)
 */

/**
 * DONE 15:: Steps should return a single struct
 * Right now they return mutliple objects which might make it hard to extend the API
 * later on.
 */

/**
 * DONE 16:: clean up panics and add some great error formatting.
 */

/**
 * DONE 17:: Test logging should be enabled in production as well.
 */

/**
 * DONE::18:: Clean up all documentation.
 *
 * Remaining: internal.go, paths.go, testing.go
 */

/**
 * 19:: Add startup flag to disable step json logging during tests.
 */

/**
 * 20:: To Confirm: Write a test to verify (or just restructure the code so that...) both
 * the test and production code are always producing the same commands.
 **/
