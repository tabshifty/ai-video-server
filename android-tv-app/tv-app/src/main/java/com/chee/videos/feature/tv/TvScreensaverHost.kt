package com.chee.videos.feature.tv

import android.content.Context
import android.content.ContextWrapper
import android.app.Activity
import android.graphics.Bitmap
import android.graphics.drawable.BitmapDrawable
import androidx.activity.compose.BackHandler
import androidx.compose.animation.Crossfade
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import coil.imageLoader
import coil.request.ImageRequest
import coil.request.SuccessResult
import coil.size.Scale
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.TvScreensaverImage
import com.chee.videos.core.model.TvScreensaverSettings
import com.chee.videos.core.repository.VideoRepository
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.util.UrlBuilder
import com.google.gson.JsonParser
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import java.util.Base64
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import okhttp3.HttpUrl.Companion.toHttpUrl

@HiltViewModel
class TvScreensaverViewModel @Inject constructor(
    store: AppPreferencesStore,
    repository: VideoRepository,
    val controller: TvScreensaverController,
    @ApplicationContext private val context: Context,
) : ViewModel() {
    val settings = store.tvScreensaverSettingsFlow.stateIn(viewModelScope, SharingStarted.Eagerly, null)
    private val _frame = MutableStateFlow<TvScreensaverFrame<Bitmap>?>(null)
    internal val frame = _frame.asStateFlow()
    private var baseUrl = ""

    init {
        viewModelScope.launch {
            TvScreensaverSession(
                fetchImages = repository::fetchTvScreensaverImages,
                loadImage = ::loadImage,
                onFrame = { frame ->
                    _frame.value = frame
                    controller.setShowing(frame != null)
                },
            ).run(store.tvScreensaverSettingsFlow, controller.runtime)
        }
    }

    internal fun setSession(baseUrl: String?, accessToken: String?) {
        this.baseUrl = baseUrl.orEmpty()
        controller.setSession(screensaverSessionKey(baseUrl, accessToken))
    }

    private suspend fun loadImage(image: TvScreensaverImage): Bitmap? {
        val metrics = context.resources.displayMetrics
        // 两张全屏解码图需同时驻留；限制到 1080p，避免小 heap 盒子被 4K 原图挤爆。
        val width = metrics.widthPixels.coerceIn(1, 1920)
        val height = metrics.heightPixels.coerceIn(1, 1080)
        val url = (UrlBuilder.normalizeBaseUrl(baseUrl) + image.viewUrl).toHttpUrl().newBuilder()
            .setQueryParameter("w", width.toString())
            .setQueryParameter("h", height.toString())
            .setQueryParameter("fit", "contain")
            .setQueryParameter("q", "82")
            .setQueryParameter("v", image.updatedAt)
            .build()
        val result = context.imageLoader.execute(
            ImageRequest.Builder(context)
                .data(url.toString())
                .size(width, height)
                .scale(Scale.FIT)
                .allowHardware(false)
                .crossfade(false)
                .build(),
        )
        // 未注册动画 decoder；只接受静态 bitmap，无法解码的格式由轮播会话跳过。
        return ((result as? SuccessResult)?.drawable as? BitmapDrawable)?.bitmap
    }
}

internal fun screensaverSessionKey(baseUrl: String?, accessToken: String?): String? {
    if (baseUrl.isNullOrBlank() || accessToken.isNullOrBlank()) return null
    // 这里只读取已登录状态的账号身份，不代替服务端鉴权；正常 token 刷新不能重启屏保。
    val account = try {
        val payload = JsonParser.parseString(
            String(Base64.getUrlDecoder().decode(accessToken.split('.')[1]), Charsets.UTF_8),
        ).asJsonObject
        (payload.get("uid") ?: payload.get("sub"))?.asString?.takeIf { it.isNotBlank() }
    } catch (_: Exception) {
        null
    }
    return "${UrlBuilder.normalizeBaseUrl(baseUrl)}\n${account ?: accessToken}"
}

@Composable
fun TvScreensaverHost(
    baseUrl: String?,
    accessToken: String?,
    viewModel: TvScreensaverViewModel = hiltViewModel(),
    content: @Composable () -> Unit,
) {
    val controller = viewModel.controller
    val runtime by controller.runtime.collectAsStateWithLifecycle()
    val showing by controller.showing.collectAsStateWithLifecycle()
    val frame by viewModel.frame.collectAsStateWithLifecycle()
    val settings by viewModel.settings.collectAsStateWithLifecycle()
    val lifecycleOwner = LocalLifecycleOwner.current
    val sessionKey = screensaverSessionKey(baseUrl, accessToken)

    DisposableEffect(viewModel, sessionKey) {
        viewModel.setSession(baseUrl, accessToken)
        onDispose { viewModel.setSession(null, null) }
    }
    DisposableEffect(controller, lifecycleOwner) {
        val observer = LifecycleEventObserver { _, _ ->
            controller.setForeground(lifecycleOwner.lifecycle.currentState.isAtLeast(Lifecycle.State.RESUMED))
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        controller.setForeground(lifecycleOwner.lifecycle.currentState.isAtLeast(Lifecycle.State.RESUMED))
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
            controller.setForeground(false)
        }
    }

    val activeSession = sessionKey != null && runtime.sessionKey == sessionKey
    KeepScreenOnEffect(enabled = activeSession && runtime.foreground && settings?.enabled == true)
    CompositionLocalProvider(LocalTvScreensaverController provides controller) {
        Box(modifier = Modifier.fillMaxSize()) {
            content()
            val displayed = frame
            if (activeSession && showing && displayed != null) {
                TvScreensaverSystemBarsEffect()
                // 不申请焦点、不替换导航，底层暂停播放器与原焦点原位保留。
                Crossfade(
                    targetState = displayed,
                    animationSpec = tween(400),
                    label = "图库屏保切换",
                    modifier = Modifier.fillMaxSize().background(Color.Black),
                ) { slide ->
                    val bitmap = remember(slide.content) { slide.content.asImageBitmap() }
                    Image(
                        bitmap = bitmap,
                        contentDescription = "屏保图片",
                        contentScale = ContentScale.Fit,
                        modifier = Modifier.fillMaxSize(),
                    )
                }
            }
        }
        // 覆盖 Compose/系统返回回调；普通遥控按键仍在 Activity 层完整消费。
        BackHandler(enabled = activeSession && showing, onBack = controller::userActivity)
    }
}

@Composable
private fun TvScreensaverSystemBarsEffect() {
    val activity = LocalContext.current.screensaverActivity()
    DisposableEffect(activity) {
        val window = activity?.window
        val decor = window?.decorView
        val controller = if (window != null && decor != null) WindowCompat.getInsetsController(window, decor) else null
        val insets = decor?.let { androidx.core.view.ViewCompat.getRootWindowInsets(it) }
        var restoreTypes = 0
        if (insets?.isVisible(WindowInsetsCompat.Type.statusBars()) == true) restoreTypes = restoreTypes or WindowInsetsCompat.Type.statusBars()
        if (insets?.isVisible(WindowInsetsCompat.Type.navigationBars()) == true) restoreTypes = restoreTypes or WindowInsetsCompat.Type.navigationBars()
        val previousBehavior = controller?.systemBarsBehavior
        controller?.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
        controller?.hide(WindowInsetsCompat.Type.systemBars())
        onDispose {
            if (restoreTypes != 0) controller?.show(restoreTypes)
            if (previousBehavior != null) controller.systemBarsBehavior = previousBehavior
        }
    }
}

private tailrec fun Context.screensaverActivity(): Activity? = when (this) {
    is Activity -> this
    is ContextWrapper -> baseContext.screensaverActivity()
    else -> null
}
