package com.paulcoding.androidtemplate

import com.paulcoding.androidtemplate.di.appModule
import com.paulcoding.androidtemplate.di.platformModule
import org.koin.core.KoinApplication
import org.koin.core.context.startKoin
import org.koin.core.context.stopKoin
import org.koin.dsl.KoinAppDeclaration

fun initKoin(config: KoinAppDeclaration = {}): KoinApplication {
    stopKoin()
    return startKoin {
        config()
        modules(appModule, platformModule)
    }
}
