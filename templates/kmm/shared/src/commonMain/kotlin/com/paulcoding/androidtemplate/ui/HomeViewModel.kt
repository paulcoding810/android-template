package com.paulcoding.androidtemplate.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.paulcoding.androidtemplate.data.Post
import com.paulcoding.androidtemplate.data.PostRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

data class HomeUiState(
    val posts: List<Post> = emptyList(),
    val isLoading: Boolean = false,
    val errorMessage: String? = null,
)

class HomeViewModel(
    private val repository: PostRepository,
) : ViewModel() {
    private val isLoading = MutableStateFlow(false)
    private val errorMessage = MutableStateFlow<String?>(null)

    val uiState: StateFlow<HomeUiState> = combine(
        repository.observePosts(),
        isLoading,
        errorMessage,
    ) { posts, loading, error ->
        HomeUiState(
            posts = posts,
            isLoading = loading,
            errorMessage = error,
        )
    }.catch { throwable ->
        emit(HomeUiState(errorMessage = throwable.message))
    }.stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5_000),
        initialValue = HomeUiState(isLoading = true),
    )

    init {
        refresh()
    }

    fun refresh() {
        viewModelScope.launch {
            isLoading.value = true
            errorMessage.value = null
            runCatching {
                repository.refreshPosts()
            }.onFailure { throwable ->
                errorMessage.value = throwable.message ?: "Unable to fetch posts."
            }
            isLoading.value = false
        }
    }
}
