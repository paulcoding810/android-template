package com.paulcoding.androidtemplate.feature.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.paulcoding.androidtemplate.core.domain.repository.HomeRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

class HomesViewModel(homeRepository: HomeRepository) : ViewModel() {
    private var _uiState = MutableStateFlow(HomeState())
    val uiState = _uiState.asStateFlow()

    init {
        viewModelScope.launch {
            _uiState.update { it.copy(isFetchingPosts = true, error = null) }
            homeRepository.getPost()
                .onSuccess { data ->
                    println("data=$data")
                    _uiState.update { it.copy(posts = data) }
                }
                .onFailure { error ->
                    error.printStackTrace()
                    _uiState.update { it.copy(error = error) }
                }
            _uiState.update { it.copy(isFetchingPosts = false) }
        }
    }
}