package com.paulcoding.androidtemplate.core.network.model

import com.paulcoding.androidtemplate.core.model.Post
import kotlinx.serialization.Serializable

@Serializable
data class NetworkPost(
    val body: String,
    val id: Int,
    val title: String,
    val userId: Int
)

fun NetworkPost.asExternalModel() = Post(
    body = body,
    id = id,
    title = title,
    userId = userId,
)