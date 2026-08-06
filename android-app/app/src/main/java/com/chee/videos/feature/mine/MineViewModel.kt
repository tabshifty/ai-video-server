package com.chee.videos.feature.mine

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.PhoneContentSourcePage
import com.chee.videos.core.model.UserProfileDto
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.model.loadPhoneContentBatch
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

enum class MineSection(
    val source: String,
    val title: String,
) {
    HISTORY("history", "历史记录"),
    FAVORITE("favorite", "我的收藏"),
    LIKE("like", "我的喜欢"),
}

data class MineVideoItem(
    val videoId: String,
    val title: String,
    val type: String,
    val thumbnailPath: String?,
    val duration: Int,
    val subtitle: String? = null,
    val progress: Float? = null,
)

data class MineListState(
    val loading: Boolean = false,
    val refreshing: Boolean = false,
    val loadingMore: Boolean = false,
    val loaded: Boolean = false,
    val page: Int = 0,
    val hasMore: Boolean = true,
    val items: List<MineVideoItem> = emptyList(),
    val errorMessage: String? = null,
)

data class MineUiState(
    val selectedSection: MineSection = MineSection.HISTORY,
    val profileLoading: Boolean = false,
    val profile: UserProfileDto? = null,
    val profileErrorMessage: String? = null,
    val history: MineListState = MineListState(),
    val favorite: MineListState = MineListState(),
    val like: MineListState = MineListState(),
)

@HiltViewModel
class MineViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(MineUiState())
    val uiState: StateFlow<MineUiState> = _uiState.asStateFlow()

    init {
        refreshProfile()
        loadSection(MineSection.HISTORY)
    }

    fun selectSection(section: MineSection) {
        _uiState.update { it.copy(selectedSection = section) }
        loadSection(section, forceRefresh = true)
    }

    fun refreshCurrentSection() {
        loadSection(_uiState.value.selectedSection, forceRefresh = true)
    }

    fun loadNextPageForCurrentSection() {
        val section = _uiState.value.selectedSection
        val state = listStateFor(section)
        if (!state.loaded) {
            loadSection(section)
            return
        }
        loadSection(section, append = true)
    }

    fun refreshProfile() {
        val current = _uiState.value
        if (current.profileLoading) {
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(profileLoading = true, profileErrorMessage = null) }
            videoRepository.fetchUserProfile()
                .onSuccess { profile ->
                    _uiState.update {
                        it.copy(
                            profileLoading = false,
                            profile = profile,
                            profileErrorMessage = null,
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            profileLoading = false,
                            profileErrorMessage = err.message ?: "用户信息加载失败",
                        )
                    }
                }
        }
    }

    private fun loadSection(
        section: MineSection,
        forceRefresh: Boolean = false,
        append: Boolean = false,
    ) {
        val current = listStateFor(section)
        if (current.loading || current.refreshing || current.loadingMore) {
            return
        }
        if (append && !current.hasMore) {
            return
        }

        val page = if (append) current.page + 1 else 1
        val showInitialLoading = !append && current.items.isEmpty()

        updateListState(section) { state ->
            when {
                append -> state.copy(loadingMore = true, errorMessage = null)
                showInitialLoading -> state.copy(loading = true, errorMessage = null)
                forceRefresh || state.loaded -> state.copy(refreshing = true, errorMessage = null)
                else -> state.copy(loading = true, errorMessage = null)
            }
        }

        viewModelScope.launch {
            loadPhoneContentBatch(
                firstPage = page,
                minimumVisibleItems = 1,
                typeOf = MineVideoItem::type,
                fetchPage = { sourcePage -> fetchSectionPage(section, sourcePage) },
            ).onSuccess { batch ->
                onSectionLoadSuccess(
                    section = section,
                    page = batch.lastLoadedPage,
                    incoming = batch.items,
                    hasMore = batch.hasMore,
                    append = append,
                    trailingErrorMessage = batch.trailingError?.message,
                )
            }.onFailure { err ->
                onSectionLoadFailure(section, err)
            }
        }
    }

    private suspend fun fetchSectionPage(
        section: MineSection,
        page: Int,
    ): Result<PhoneContentSourcePage<MineVideoItem>> {
        return when (section) {
            MineSection.HISTORY -> videoRepository.fetchContinueHistory(page = page, limit = PAGE_SIZE)
                .map { payload ->
                    PhoneContentSourcePage(
                        items = payload.items.map { history ->
                            MineVideoItem(
                                videoId = history.videoId,
                                title = history.title,
                                type = history.type,
                                thumbnailPath = history.thumbnailPath,
                                duration = history.duration,
                                subtitle = "已观看 ${history.watchSeconds} 秒",
                                progress = history.progress.coerceIn(0f, 1f),
                            )
                        },
                        hasMore = hasMoreSourcePages(
                            page = payload.page,
                            pageSize = payload.pageSize,
                            totalCount = payload.totalCount,
                        ),
                    )
                }

            MineSection.FAVORITE -> videoRepository.fetchFavoritedVideos(page = page, pageSize = PAGE_SIZE)
                .map { payload ->
                    PhoneContentSourcePage(
                        items = payload.items.map { row -> row.toMineVideoItem() },
                        hasMore = hasMoreSourcePages(
                            page = payload.page,
                            pageSize = payload.pageSize,
                            totalCount = payload.totalCount,
                        ),
                    )
                }

            MineSection.LIKE -> videoRepository.fetchLikedVideos(page = page, pageSize = PAGE_SIZE)
                .map { payload ->
                    PhoneContentSourcePage(
                        items = payload.items.map { row -> row.toMineVideoItem() },
                        hasMore = hasMoreSourcePages(
                            page = payload.page,
                            pageSize = payload.pageSize,
                            totalCount = payload.totalCount,
                        ),
                    )
                }
        }
    }

    private fun VideoListItemDto.toMineVideoItem(): MineVideoItem {
        return MineVideoItem(
            videoId = id,
            title = title,
            type = type,
            thumbnailPath = thumbnailPath,
            duration = duration,
            subtitle = "${typeLabel(type)} · $duration 秒",
        )
    }

    private fun onSectionLoadSuccess(
        section: MineSection,
        page: Int,
        incoming: List<MineVideoItem>,
        hasMore: Boolean,
        append: Boolean,
        trailingErrorMessage: String?,
    ) {
        updateListState(section) { state ->
            val mergedItems = if (append) mergeByVideoId(state.items, incoming) else incoming
            state.copy(
                loading = false,
                refreshing = false,
                loadingMore = false,
                loaded = true,
                page = page,
                hasMore = hasMore,
                items = mergedItems,
                errorMessage = trailingErrorMessage,
            )
        }
    }

    private fun onSectionLoadFailure(section: MineSection, err: Throwable) {
        handleAuthError(err)
        updateListState(section) { state ->
            state.copy(
                loading = false,
                refreshing = false,
                loadingMore = false,
                loaded = true,
                errorMessage = err.message ?: "加载失败",
            )
        }
    }

    private fun listStateFor(section: MineSection): MineListState {
        return when (section) {
            MineSection.HISTORY -> _uiState.value.history
            MineSection.FAVORITE -> _uiState.value.favorite
            MineSection.LIKE -> _uiState.value.like
        }
    }

    private fun updateListState(section: MineSection, transform: (MineListState) -> MineListState) {
        _uiState.update { state ->
            when (section) {
                MineSection.HISTORY -> state.copy(history = transform(state.history))
                MineSection.FAVORITE -> state.copy(favorite = transform(state.favorite))
                MineSection.LIKE -> state.copy(like = transform(state.like))
            }
        }
    }

    private fun mergeByVideoId(existing: List<MineVideoItem>, incoming: List<MineVideoItem>): List<MineVideoItem> {
        if (existing.isEmpty()) {
            return incoming
        }
        if (incoming.isEmpty()) {
            return existing
        }

        val merged = existing.toMutableList()
        val seen = existing.map { it.videoId }.toMutableSet()
        incoming.forEach { row ->
            if (seen.add(row.videoId)) {
                merged.add(row)
            }
        }
        return merged
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is AuthExpiredException) {
            viewModelScope.launch {
                authRepository.logoutLocal()
            }
        }
    }

    private fun typeLabel(type: String): String {
        return when (type) {
            "short" -> "短视频"
            "av" -> "AV"
            else -> "视频"
        }
    }

    private fun hasMoreSourcePages(page: Int, pageSize: Int, totalCount: Int): Boolean {
        val safePage = page.coerceAtLeast(1)
        val safePageSize = pageSize.coerceAtLeast(1)
        return safePage * safePageSize < totalCount.coerceAtLeast(0)
    }

    private companion object {
        const val PAGE_SIZE = 20
    }
}
