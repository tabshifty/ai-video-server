package com.chee.videos.feature.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.repository.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class LoginUiState(
    val identity: String = "",
    val password: String = "",
    val loading: Boolean = false,
    val errorMessage: String? = null,
)

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    fun updateIdentity(value: String) {
        _uiState.update { it.copy(identity = value, errorMessage = null) }
    }

    fun updatePassword(value: String) {
        _uiState.update { it.copy(password = value, errorMessage = null) }
    }

    fun login() {
        val identity = _uiState.value.identity.trim()
        val password = _uiState.value.password
        if (identity.isBlank() || password.isBlank()) {
            _uiState.update { it.copy(errorMessage = "请输入账号和密码") }
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(loading = true, errorMessage = null) }
            authRepository.login(identity, password)
                .onSuccess {
                    _uiState.update { it.copy(loading = false, password = "") }
                }
                .onFailure { err ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            errorMessage = err.message ?: "登录失败",
                        )
                    }
                }
        }
    }
}
