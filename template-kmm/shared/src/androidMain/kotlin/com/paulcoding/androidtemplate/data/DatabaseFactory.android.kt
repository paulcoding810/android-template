package com.paulcoding.androidtemplate.data

import android.content.Context
import androidx.room.Room
import androidx.room.RoomDatabase

actual class DatabaseFactory(
    private val context: Context,
) {
    actual fun createBuilder(): RoomDatabase.Builder<AppDatabase> {
        return Room.databaseBuilder(
            context = context.applicationContext,
            klass = AppDatabase::class.java,
            name = DatabaseName,
        )
    }
}
