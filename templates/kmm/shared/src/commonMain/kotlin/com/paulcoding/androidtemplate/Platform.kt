package com.paulcoding.androidtemplate

interface Platform {
    val name: String
}

expect fun getPlatform(): Platform