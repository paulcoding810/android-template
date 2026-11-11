package com.paulcoding.androidtemplate.jvm

import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import com.paulcoding.androidtemplate.App
import com.paulcoding.androidtemplate.initKoin

fun main() {
    initKoin()

    application {
        Window(
            onCloseRequest = ::exitApplication,
            title = "Android Template",
        ) {
            App()
        }
    }
}
