package com.chee.videos.core.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.AppRootState
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.ServerRepository
import com.chee.videos.core.util.UrlBuilder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

@HiltViewModel
class AppRootViewModel @Inject constructor(
    private val store: AppPreferencesStore,
    private val authRepository: AuthRepository,
    private val serverRepository: ServerRepository,
) : ViewModel() {

    val appState: StateFlow<AppRootState> = combine(
        store.activeBaseUrlFlow,
        store.accessTokenFlow,
    ) { baseUrl, accessToken ->
        when {
            baseUrl.isNullOrBlank() -> AppRootState.NeedServer
            accessToken.isNullOrBlank() -> AppRootState.NeedLogin
            else -> AppRootState.Ready(baseUrl, accessToken)
        }
    }.stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5000),
        initialValue = AppRootState.Loading,
    )

    fun switchToServerSelection() {
        viewModelScope.launch {
            store.clearActiveServerAndTokens()
        }
    }

    fun logout() {
        viewModelScope.launch {
            authRepository.logoutLocal()
        }
    }

    fun applyTvAuthServer(serverBaseUrl: String?) {
        val normalized = serverBaseUrl?.let(UrlBuilder::normalizeBaseUrl).orEmpty()
        if (normalized.isBlank()) {
            return
        }
        viewModelScope.launch {
            val current = store.readActiveBaseUrl().orEmpty()
            if (current == normalized) {
                return@launch
            }
            serverRepository.activateEndpoint(normalized, clearTokens = true)
        }
    }
}
