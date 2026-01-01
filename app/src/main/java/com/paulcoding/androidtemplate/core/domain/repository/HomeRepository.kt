package com.paulcoding.androidtemplate.core.domain.repository

import com.paulcoding.androidtemplate.core.model.Post
import com.paulcoding.androidtemplate.core.network.datasource.KtorPostDataSource
import com.paulcoding.androidtemplate.core.network.model.NetworkPost
import com.paulcoding.androidtemplate.core.network.model.asExternalModel
import com.paulcoding.androidtemplate.core.network.util.apiCall

class HomeRepository(private val dataSource: KtorPostDataSource) {
    suspend fun getPost(): Result<List<Post>> = apiCall { dataSource.getPosts().map(NetworkPost::asExternalModel) }
}
