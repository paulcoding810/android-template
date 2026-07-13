package com.paulcoding.androidtemplate.core.network.datasource

import com.paulcoding.androidtemplate.core.network.model.NetworkPost

interface PostDataSource {
    suspend fun getPosts(): List<NetworkPost>
}