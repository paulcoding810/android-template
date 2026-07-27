package com.paulcoding.androidtemplate.di

import com.paulcoding.androidtemplate.data.DatabaseFactory
import org.koin.core.module.Module
import org.koin.dsl.module

actual val platformModule: Module = module {
    single { DatabaseFactory() }
}
