package com.paulcoding.androidtemplate.data

import androidx.room.RoomDatabase
import androidx.sqlite.driver.bundled.BundledSQLiteDriver
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.IO

internal const val DatabaseName = "android-template.db"

expect class DatabaseFactory {
    fun createBuilder(): RoomDatabase.Builder<AppDatabase>
}

fun createDatabase(factory: DatabaseFactory): AppDatabase {
    return factory.createBuilder()
        .setDriver(BundledSQLiteDriver())
        .setQueryCoroutineContext(Dispatchers.IO)
        .build()
}
