package com.chee.videos.feature.tv

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.selection.selectableGroup
import androidx.compose.foundation.selection.toggleable
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.TvScreensaverSettings
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.tvFocusableScaleOnly
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

@HiltViewModel
class TvScreensaverSettingsViewModel @Inject constructor(private val store: AppPreferencesStore) : ViewModel() {
    val settings = store.tvScreensaverSettingsFlow.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000L), null)
    private val _error = MutableStateFlow<String?>(null)
    val error = _error.asStateFlow()

    fun setEnabled(enabled: Boolean) = save { store.saveTvScreensaverEnabled(enabled) }
    fun setIdleMinutes(minutes: Int) = save { store.saveTvScreensaverIdleMinutes(minutes) }
    fun setSlideSeconds(seconds: Int) = save { store.saveTvScreensaverSlideSeconds(seconds) }

    private fun save(action: suspend () -> Unit) {
        viewModelScope.launch {
            try {
                action()
                _error.value = null
            } catch (err: CancellationException) {
                throw err
            } catch (_: Exception) {
                _error.value = "屏保设置保存失败，请重试"
            }
        }
    }
}

@Composable
internal fun TvScreensaverSettingsSection(viewModel: TvScreensaverSettingsViewModel = hiltViewModel()) {
    val loaded by viewModel.settings.collectAsStateWithLifecycle()
    val error by viewModel.error.collectAsStateWithLifecycle()
    Text("屏保设置", color = AppChrome.TextMuted, style = MaterialTheme.typography.bodyMedium,
        modifier = Modifier.padding(top = 10.dp))
    val settings = loaded
    if (settings == null) {
        Text("正在读取屏保设置", color = AppChrome.TextMuted)
        return
    }
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.SurfaceShape,
        modifier = Modifier.fillMaxWidth().heightIn(min = 76.dp)
            .tvFocusableScaleOnly(focusedScale = 1.02f)
            .toggleable(value = settings.enabled, role = Role.Switch, onValueChange = viewModel::setEnabled),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Text("开启屏保", color = AppChrome.TextPrimary, style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold)
                Text("闲置时自动播放图库图片，等待和轮播期间保持亮屏", color = AppChrome.TextMuted,
                    style = MaterialTheme.typography.bodySmall)
            }
            Switch(checked = settings.enabled, onCheckedChange = null)
        }
    }
    TvScreensaverDurationRow(
        title = "无操作后进入屏保", options = TvScreensaverSettings.idleMinuteOptions,
        selected = settings.idleMinutes, unit = "分钟", enabled = settings.enabled,
        onSelect = viewModel::setIdleMinutes,
    )
    TvScreensaverDurationRow(
        title = "每张图片停留时间", options = TvScreensaverSettings.slideSecondOptions,
        selected = settings.slideSeconds, unit = "秒", enabled = settings.enabled,
        onSelect = viewModel::setSlideSeconds,
    )
    error?.let { Text(it, color = AppChrome.TextPrimary, style = MaterialTheme.typography.bodySmall) }
}

@Composable
private fun TvScreensaverDurationRow(
    title: String,
    options: List<Int>,
    selected: Int,
    unit: String,
    enabled: Boolean,
    onSelect: (Int) -> Unit,
) {
    Column(
        modifier = Modifier.fillMaxWidth().alpha(if (enabled) 1f else 0.45f)
            .background(AppChrome.SurfaceElevated, AppChrome.SurfaceShape)
            .padding(horizontal = 18.dp, vertical = 16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text(title, color = AppChrome.TextPrimary, style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold)
        Row(modifier = Modifier.selectableGroup(), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
            options.forEach { value ->
                val isSelected = selected == value
                Surface(
                    color = if (isSelected) AppChrome.Accent.copy(alpha = 0.92f) else AppChrome.Surface.copy(alpha = 0.72f),
                    shape = AppChrome.ChipShape,
                    modifier = Modifier.width(80.dp).height(44.dp)
                        .tvFocusableScaleOnly(enabled = enabled, focusedScale = 1.04f)
                        .selectable(selected = isSelected, enabled = enabled, role = Role.RadioButton,
                            onClick = { onSelect(value) }),
                ) {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        Text("$value $unit", color = if (isSelected) AppChrome.Canvas else AppChrome.TextPrimary,
                            style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}
