package com.paulcoding.androidtemplate

import com.paulcoding.androidtemplate.feature.home.HomesViewModel
import org.koin.core.module.dsl.viewModelOf
import org.koin.dsl.module

val viewModelModule = module {
    viewModelOf(::HomesViewModel)
}