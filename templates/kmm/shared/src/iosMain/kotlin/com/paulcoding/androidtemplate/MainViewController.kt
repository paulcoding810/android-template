package com.paulcoding.androidtemplate

import androidx.compose.ui.window.ComposeUIViewController
import platform.UIKit.UIViewController

fun MainViewController(): UIViewController {
    initKoin()
    return ComposeUIViewController {
        App()
    }
}
