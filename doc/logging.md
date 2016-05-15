# Logging

Alda uses [timbre](https://github.com/ptaoussanis/timbre) for logging. Every note event, attribute change, etc. is logged at the DEBUG level, which can be useful for debugging purposes.

The default logging level is WARN, so by default, you will not see these debug-level logs; you will only see warnings and errors.

To override this setting (e.g. for development and debugging), you can set the `TIMBRE_LEVEL` environment variable.

To see debug logs, for example, you can do this:

    export TIMBRE_LEVEL=debug

When running tests via `boot test` and troubleshooting a failing test, it may be helpful to use debug-level logging by running `TIMBRE_LEVEL=debug boot test`.

