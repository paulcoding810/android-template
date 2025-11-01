package com.paulcoding.androidtemplate.feature.home

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import org.koin.androidx.compose.koinViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    viewModel: HomesViewModel = koinViewModel(),
) {
    val state = viewModel.uiState.collectAsState()
    Scaffold(topBar = {
        TopAppBar(title = {
            Text("Posts")
        })
    }) { innerPadding ->
        HomeContent(modifier = Modifier.padding(innerPadding), state = state.value)
    }
}

@Composable
fun HomeContent(modifier: Modifier = Modifier, state: HomeState) {
    LazyColumn(modifier = modifier, verticalArrangement = Arrangement.spacedBy(8.dp)) {
        items(items = state.posts, key = { it.id }) { post ->
            Column {
                Text(post.title, style = MaterialTheme.typography.titleSmall)
                Text(post.body)
            }
        }
    }
}