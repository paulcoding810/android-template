package com.paulcoding.androidtemplate.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalInspectionMode
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.paulcoding.androidtemplate.data.Post
import org.koin.core.component.KoinComponent
import org.koin.core.component.inject

@Composable
fun HomeRoute() {
    if (LocalInspectionMode.current) {
        HomeScreen(
            uiState = HomeUiState(posts = previewPosts),
            onRefreshClick = {},
        )
        return
    }

    val repository = remember { HomeViewModelFactory.repository }
    val viewModel = viewModel { HomeViewModel(repository) }
    val uiState by viewModel.uiState.collectAsState()

    HomeScreen(
        uiState = uiState,
        onRefreshClick = viewModel::refresh,
    )
}

private object HomeViewModelFactory : KoinComponent {
    val repository by inject<com.paulcoding.androidtemplate.data.PostRepository>()
}

private val previewPosts = listOf(
    Post(
        userId = 1,
        id = 1,
        title = "hello from jsonplaceholder",
        body = "This sample post keeps Compose previews independent from Koin and network calls.",
    ),
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    uiState: HomeUiState,
    onRefreshClick: () -> Unit,
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Posts") },
                actions = {
                    Button(
                        onClick = onRefreshClick,
                        enabled = !uiState.isLoading,
                    ) {
                        Text("Refresh")
                    }
                },
            )
        },
    ) { paddingValues ->
        Surface(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
            color = MaterialTheme.colorScheme.background,
        ) {
            when {
                uiState.posts.isEmpty() && uiState.isLoading -> LoadingContent()
                uiState.posts.isEmpty() && uiState.errorMessage != null -> EmptyErrorContent(
                    message = uiState.errorMessage,
                    onRefreshClick = onRefreshClick,
                )
                else -> PostListContent(uiState)
            }
        }
    }
}

@Composable
private fun LoadingContent() {
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center,
    ) {
        CircularProgressIndicator()
    }
}

@Composable
private fun EmptyErrorContent(
    message: String,
    onRefreshClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = "Could not load posts",
            style = MaterialTheme.typography.titleLarge,
            fontWeight = FontWeight.Bold,
        )
        Spacer(Modifier.height(8.dp))
        Text(message)
        Spacer(Modifier.height(16.dp))
        Button(onClick = onRefreshClick) {
            Text("Try again")
        }
    }
}

@Composable
private fun PostListContent(uiState: HomeUiState) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Column {
                    Text(
                        text = "Home",
                        style = MaterialTheme.typography.headlineMedium,
                        fontWeight = FontWeight.Bold,
                    )
                    Text("${uiState.posts.size} posts from JSONPlaceholder")
                }
                if (uiState.isLoading) {
                    CircularProgressIndicator()
                }
            }
        }

        uiState.errorMessage?.let { message ->
            item {
                Text(
                    text = message,
                    color = MaterialTheme.colorScheme.error,
                )
            }
        }

        items(uiState.posts, key = Post::id) { post ->
            PostCard(post)
        }
    }
}

@Composable
private fun PostCard(post: Post) {
    Card(
        modifier = Modifier.fillMaxWidth(),
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(
                text = post.title.replaceFirstChar { it.titlecase() },
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
            )
            Text(post.body)
        }
    }
}
