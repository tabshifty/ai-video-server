package com.chee.videos.feature.imagecollections

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ImageCollectionDetailDto
import com.chee.videos.core.model.ImageCollectionListItemDto
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ImageCollectionsUiState(
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val loaded: Boolean = false,
    val page: Int = 0,
    val totalCount: Int = 0,
    val items: List<ImageCollectionListItemDto> = emptyList(),
    val errorMessage: String? = null,
)

data class ImageCollectionViewerUiState(
    val loading: Boolean = true,
    val detail: ImageCollectionDetailDto? = null,
    val errorMessage: String? = null,
)

@HiltViewModel
class ImageCollectionsViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(ImageCollectionsUiState())
    val uiState: StateFlow<ImageCollectionsUiState> = _uiState.asStateFlow()

    fun load(force: Boolean = false) {
        val state = _uiState.value
        if (state.loading || (state.loaded && !force)) {
            return
        }
        _uiState.update {
            it.copy(
                loading = true,
                errorMessage = null,
                page = if (force) 0 else it.page,
                totalCount = if (force) 0 else it.totalCount,
                items = if (force) emptyList() else it.items,
            )
        }
        loadPage(page = 1, append = false)
    }

    fun retry() {
        _uiState.update {
            it.copy(
                loading = true,
                loadingMore = false,
                errorMessage = null,
                page = 0,
                totalCount = 0,
                items = emptyList(),
            )
        }
        loadPage(page = 1, append = false)
    }

    fun loadMoreIfNeeded(currentIndex: Int) {
        val state = _uiState.value
        if (state.loading || state.loadingMore || state.items.isEmpty()) {
            return
        }
        if (currentIndex < state.items.size - 6) {
            return
        }
        if (state.totalCount > 0 && state.items.size >= state.totalCount) {
            return
        }
        loadPage(page = state.page + 1, append = true)
    }

    private fun loadPage(page: Int, append: Boolean) {
        viewModelScope.launch {
            if (append) {
                _uiState.update { it.copy(loadingMore = true, errorMessage = null) }
            }
            videoRepository.fetchImageCollections(page = page, pageSize = 24)
                .onSuccess { payload ->
                    _uiState.update {
                        val mergedItems = if (append) {
                            (it.items + payload.items).distinctBy { row -> row.id }
                        } else {
                            payload.items.distinctBy { row -> row.id }
                        }
                        it.copy(
                            loading = false,
                            loadingMore = false,
                            loaded = true,
                            page = payload.page.coerceAtLeast(page),
                            totalCount = payload.totalCount.coerceAtLeast(mergedItems.size),
                            items = mergedItems,
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            loading = false,
                            loadingMore = false,
                            loaded = true,
                            errorMessage = err.message ?: "图集加载失败",
                        )
                    }
                }
        }
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is AuthExpiredException) {
            viewModelScope.launch {
                authRepository.logoutLocal()
            }
        }
    }
}

@HiltViewModel
class ImageCollectionViewerViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val videoRepository: VideoRepository,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val collectionId: String = checkNotNull(savedStateHandle["collectionId"])

    private val _uiState = MutableStateFlow(ImageCollectionViewerUiState())
    val uiState: StateFlow<ImageCollectionViewerUiState> = _uiState.asStateFlow()

    init {
        load()
    }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true, errorMessage = null) }
            videoRepository.fetchImageCollectionDetail(collectionId)
                .onSuccess { detail ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            detail = detail,
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            loading = false,
                            errorMessage = err.message ?: "图集详情加载失败",
                        )
                    }
                }
        }
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is AuthExpiredException) {
            viewModelScope.launch {
                authRepository.logoutLocal()
            }
        }
    }
}
