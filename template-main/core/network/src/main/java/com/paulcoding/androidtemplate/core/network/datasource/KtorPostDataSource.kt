package com.paulcoding.androidtemplate.core.network.datasource

import com.paulcoding.androidtemplate.core.network.model.NetworkPost
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.get

class KtorPostDataSource(private val httpClient: HttpClient) : PostDataSource {
    override suspend fun getPosts() = httpClient.get("posts").body<List<NetworkPost>>()
}