package com.paulcoding.androidtemplate.data

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

class PostRepository(
    private val api: PostApi,
    private val dao: PostDao,
) {
    fun observePosts(): Flow<List<Post>> {
        return dao.observePosts().map { posts -> posts.map(PostEntity::toPost) }
    }

    suspend fun refreshPosts() {
        dao.upsertPosts(api.fetchPosts().map(Post::toEntity))
    }
}
