package com.paulcoding.androidtemplate.data

import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.get

class PostApi(
    private val httpClient: HttpClient,
) {
    suspend fun fetchPosts(): List<Post> {
        return httpClient.get("https://jsonplaceholder.typicode.com/posts").body()
    }
}
