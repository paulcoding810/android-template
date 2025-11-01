package com.paulcoding.androidtemplate

import android.app.Application
import com.paulcoding.androidtemplate.core.data.dataModule
import com.paulcoding.androidtemplate.core.domain.repositoryModule
import com.paulcoding.androidtemplate.core.network.di.networkModule
import org.koin.android.ext.koin.androidContext
import org.koin.core.context.GlobalContext

class App : Application() {
    override fun onCreate() {
        super.onCreate()

        GlobalContext.startKoin {
            androidContext(this@App)
            modules(
                listOf(
                    dataModule,
                    networkModule,
                    repositoryModule,
                    viewModelModule,
                )
            )
        }
    }
}