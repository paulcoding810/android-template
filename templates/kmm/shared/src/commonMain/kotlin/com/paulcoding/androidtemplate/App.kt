package com.paulcoding.androidtemplate

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeContentPadding
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import com.paulcoding.androidtemplate.ui.HomeRoute

@Composable
@Preview
fun App() {
    MaterialTheme {
        var selectedRoute by remember { mutableStateOf(AppRoute.Home) }

        Scaffold(
            modifier = Modifier.safeContentPadding(),
            bottomBar = {
                NavigationBar {
                    AppRoute.entries.forEach { route ->
                        NavigationBarItem(
                            selected = selectedRoute == route,
                            onClick = { selectedRoute = route },
                            icon = { Text(route.icon) },
                            label = { Text(route.label) },
                        )
                    }
                }
            },
        ) { paddingValues ->
            Box(Modifier.padding(paddingValues)) {
                when (selectedRoute) {
                    AppRoute.Home -> HomeRoute()
                    AppRoute.About -> Text("Built with Compose Multiplatform, Ktor, Koin, and Room.")
                }
            }
        }
    }
}

private enum class AppRoute(
    val label: String,
    val icon: String,
) {
    Home("Home", "🏠"),
    About("About", "ℹ️"),
}
