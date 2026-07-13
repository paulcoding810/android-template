package com.paulcoding.androidtemplate.core.network.util

import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

/**
 * Executes the given block of code on the specified dispatcher and catches any exceptions that occur.
 *
 * @param dispatcher The coroutine dispatcher to use for execution. Defaults to [Dispatchers.IO].
 * @param block The block of code to execute.
 * @return A [Result] object containing the successful result or any exception that occurred.
 */
suspend fun <T> apiCall(
    dispatcher: CoroutineDispatcher = Dispatchers.IO,
    block: suspend () -> T
): Result<T> = withContext(dispatcher) {
    runCatching { block() }
}