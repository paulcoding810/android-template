package com.paulcoding.androidtemplate.di

import com.paulcoding.androidtemplate.data.DatabaseFactory
import com.paulcoding.androidtemplate.data.PostApi
import com.paulcoding.androidtemplate.data.PostRepository
import com.paulcoding.androidtemplate.data.createDatabase
import io.ktor.client.HttpClient
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json
import org.koin.core.module.Module
import org.koin.dsl.module

expect val platformModule: Module

val appModule = module {
    single {
        HttpClient {
            install(ContentNegotiation) {
                json(
                    Json {
                        ignoreUnknownKeys = true
                    },
                )
            }
        }
    }
    single { createDatabase(get<DatabaseFactory>()) }
    single { get<com.paulcoding.androidtemplate.data.AppDatabase>().postDao() }
    single { PostApi(get()) }
    single { PostRepository(get(), get()) }
}
