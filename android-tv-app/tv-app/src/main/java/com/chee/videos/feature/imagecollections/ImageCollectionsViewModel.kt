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
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

private const val imageCollectionsSearchDebounceMillis = 350L

data class ImageCollectionsUiState(
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val loaded: Boolean = false,
    val page: Int = 0,
    val totalCount: Int = 0,
    val query: String = "",
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
    private var searchJob: Job? = null
    private var requestVersion: Long = 0

    fun load(force: Boolean = false) {
        val state = _uiState.value
        if (state.loading || (state.loaded && !force)) {
            return
        }
        if (force) {
            _uiState.update { resetImageCollectionsForQuery(it, it.query) }
        } else {
            _uiState.update { it.copy(loading = true, errorMessage = null) }
        }
        requestVersion += 1
        loadPage(page = 1, append = false, requestVersion = requestVersion)
    }

    fun retry() {
        _uiState.update { resetImageCollectionsForQuery(it, it.query) }
        requestVersion += 1
        loadPage(page = 1, append = false, requestVersion = requestVersion)
    }

    fun onQueryChanged(query: String) {
        if (query == _uiState.value.query) {
            return
        }
        _uiState.update { resetImageCollectionsForQuery(it, query) }
        searchJob?.cancel()
        requestVersion += 1
        val nextVersion = requestVersion
        searchJob = viewModelScope.launch {
            delay(imageCollectionsSearchDebounceMillis)
            loadPage(page = 1, append = false, requestVersion = nextVersion)
        }
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
        loadPage(page = state.page + 1, append = true, requestVersion = requestVersion)
    }

    private fun loadPage(page: Int, append: Boolean, requestVersion: Long) {
        viewModelScope.launch {
            if (append) {
                _uiState.update { it.copy(loadingMore = true, errorMessage = null) }
            }
            val query = normalizeImageCollectionsQuery(_uiState.value.query)
            videoRepository.fetchImageCollections(query = query, page = page, pageSize = 24)
                .onSuccess { payload ->
                    if (requestVersion != this@ImageCollectionsViewModel.requestVersion) {
                        return@onSuccess
                    }
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
                    if (requestVersion != this@ImageCollectionsViewModel.requestVersion) {
                        return@onFailure
                    }
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

internal fun normalizeImageCollectionsQuery(raw: String): String? {
    val normalized = raw.trim()
    return normalized.takeIf { it.isNotEmpty() }
}

internal fun resetImageCollectionsForQuery(
    state: ImageCollectionsUiState,
    query: String,
): ImageCollectionsUiState = state.copy(
    loading = true,
    loadingMore = false,
    loaded = false,
    page = 0,
    totalCount = 0,
    query = query,
    items = emptyList(),
    errorMessage = null,
)

internal fun imageCollectionsEmptyMessage(query: String): String {
    return if (normalizeImageCollectionsQuery(query) == null) {
        "暂无可用图片合集"
    } else {
        "没有找到相关图集"
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
