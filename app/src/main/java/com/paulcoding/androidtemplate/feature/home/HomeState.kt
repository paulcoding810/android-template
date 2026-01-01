package com.paulcoding.androidtemplate.feature.home

import com.paulcoding.androidtemplate.core.model.Post

data class HomeState(
    val isFetchingPosts: Boolean = false,
    val posts: List<Post> = emptyList(),
    val error: Throwable? = null,
)