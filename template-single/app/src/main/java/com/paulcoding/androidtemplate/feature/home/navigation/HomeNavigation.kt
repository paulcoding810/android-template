package com.paulcoding.androidtemplate.feature.home.navigation

import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavOptions
import androidx.navigation.compose.composable
import com.paulcoding.androidtemplate.core.ui.Routes
import com.paulcoding.androidtemplate.feature.home.HomeScreen

fun NavController.navigateToHome(navOptions: NavOptions) =
    navigate(route = Routes.Home, navOptions)

fun NavGraphBuilder.homeScreen() {
    composable<Routes.Home>(
        enterTransition = { fadeIn() },
        exitTransition = { fadeOut() },
        popEnterTransition = { fadeIn() },
        popExitTransition = { fadeOut() },
    ) {
        HomeScreen()
    }
}
