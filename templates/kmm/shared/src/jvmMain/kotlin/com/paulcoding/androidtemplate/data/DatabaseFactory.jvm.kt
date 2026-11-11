package com.paulcoding.androidtemplate.data

import androidx.room.Room
import androidx.room.RoomDatabase
import java.io.File

actual class DatabaseFactory {
    actual fun createBuilder(): RoomDatabase.Builder<AppDatabase> {
        val databaseFile = File(
            System.getProperty("user.home"),
            ".android-template/$DatabaseName",
        )
        databaseFile.parentFile?.mkdirs()

        return Room.databaseBuilder<AppDatabase>(name = databaseFile.absolutePath)
    }
}
