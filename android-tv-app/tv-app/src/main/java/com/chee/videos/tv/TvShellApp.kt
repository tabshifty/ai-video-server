package com.chee.videos.tv

import android.app.Activity
import android.content.Context
import android.content.ContextWrapper
import android.os.SystemClock
import androidx.activity.compose.BackHandler
import androidx.compose.animation.ExperimentalSharedTransitionApi
import androidx.compose.animation.SharedTransitionLayout
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.LocalBringIntoViewSpec
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusProperties
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.chee.videos.core.model.AppRootState
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.AppDarkColors
import com.chee.videos.core.ui.LocalTvAnimatedContentScope
import com.chee.videos.core.ui.LocalTvSharedTransitionScope
import com.chee.videos.core.ui.TvMinimalBringIntoViewSpec
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.TvTypography
import com.chee.videos.core.viewmodel.AppRootViewModel
import com.chee.videos.feature.connection.ConnectionScreen
import com.chee.videos.feature.tv.TvEpisodeArg
import com.chee.videos.feature.tv.TvLongFormDetailRoutePattern
import com.chee.videos.feature.tv.TvLongFormPlayerRoutePattern
import com.chee.videos.feature.tv.TvLongFormVideoIdArg
import com.chee.videos.feature.tv.TvLongFormVideoTypeArg
import com.chee.videos.feature.tv.TvLongFormDetailScreen
import com.chee.videos.feature.tv.TvLongFormPlayerScreen
import com.chee.videos.feature.tv.TvIptvRoute
import com.chee.videos.feature.tv.TvIptvScreen
import com.chee.videos.feature.tv.TvSeriesRoutePattern
import com.chee.videos.feature.tv.TvCatalogWallKindArg
import com.chee.videos.feature.tv.TvCatalogWallRoutePattern
import com.chee.videos.feature.tv.TvCatalogWallTitleArg
import com.chee.videos.feature.tv.TvPlayerRoutePattern
import com.chee.videos.feature.tv.TvSeasonArg
import com.chee.videos.feature.tv.TvSeriesDetailScreen
import com.chee.videos.feature.tv.TvSeriesIdArg
import com.chee.videos.feature.tv.TvSeriesPlayerScreen
import com.chee.videos.feature.tv.buildTvSeriesRoute
import com.chee.videos.feature.tv.buildTvCatalogWallRoute
import com.chee.videos.feature.tv.buildTvLongFormDetailRoute
import com.chee.videos.feature.tv.buildTvLongFormPlayerRoute
import com.chee.videos.feature.tv.buildTvPlayerRoute
import kotlinx.coroutines.delay

@Composable
fun TvShellApp(
    appRootViewModel: AppRootViewModel = hiltViewModel(),
) {
    val appState by appRootViewModel.appState.collectAsStateWithLifecycle()

    MaterialTheme(colorScheme = AppDarkColors, typography = TvTypography) {
        Surface(modifier = Modifier.fillMaxSize(), color = AppChrome.Canvas) {
            when (val state = appState) {
                AppRootState.Loading -> {
                    TvPageLoadingState(message = "正在准备 TV 首页")
                }

                AppRootState.NeedServer -> ConnectionScreen()
                AppRootState.NeedLogin -> TvPairingScreen(
                    onSwitchServer = appRootViewModel::switchToServerSelection,
                )
                is AppRootState.Ready -> TvAuthenticatedNav(
                    baseUrl = state.baseUrl,
                    accessToken = state.accessToken,
                    onLogout = appRootViewModel::logout,
                    onRepair = appRootViewModel::logout,
                    onSwitchServer = appRootViewModel::switchToServerSelection,
                )
            }
        }
    }
}

@OptIn(ExperimentalSharedTransitionApi::class, ExperimentalFoundationApi::class)
@Composable
private fun TvAuthenticatedNav(
    baseUrl: String,
    accessToken: String,
    onLogout: () -> Unit,
    onRepair: () -> Unit,
    onSwitchServer: () -> Unit,
) {
    val navController = rememberNavController()
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route
    val handleShellBack = shouldHandleTvShellBack(currentRoute)
    val handleRootExitConfirm = shouldHandleTvRootExitConfirm(currentRoute)
    val homeContentFocusRequester = remember { FocusRequester() }
    val activity = LocalContext.current.findActivity()
    var rootExitPromptAtMillis by remember { mutableStateOf<Long?>(null) }
    var showRootExitPrompt by remember { mutableStateOf(false) }

    fun handleRootBack() {
        val now = SystemClock.uptimeMillis()
        when (resolveTvRootBackAction(rootExitPromptAtMillis, now)) {
            TvRootBackAction.ShowPrompt -> {
                rootExitPromptAtMillis = now
                showRootExitPrompt = true
            }

            TvRootBackAction.Exit -> {
                rootExitPromptAtMillis = null
                showRootExitPrompt = false
                activity?.finish()
            }
        }
    }

    BackHandler(enabled = handleShellBack) {
        navController.popBackStack()
    }

    BackHandler(enabled = handleRootExitConfirm, onBack = ::handleRootBack)

    LaunchedEffect(showRootExitPrompt, rootExitPromptAtMillis) {
        val promptAt = rootExitPromptAtMillis
        if (showRootExitPrompt && promptAt != null) {
            delay(TvRootExitConfirmWindowMillis)
            if (rootExitPromptAtMillis == promptAt) {
                showRootExitPrompt = false
            }
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient),
    ) {
        // 用 TvMinimalBringIntoViewSpec 覆盖 Compose Foundation 1.7 在 leanback 设备上默认启用的
        // PivotBringIntoViewSpec，避免焦点目标已可见时仍把 LazyColumn 拉到 30% pivot 位置；
        // 否则 TV 首页 hero 大海报会在初始焦点落到底部播放按钮时被强制上滑裁切。
        CompositionLocalProvider(LocalBringIntoViewSpec provides TvMinimalBringIntoViewSpec) {
        SharedTransitionLayout(modifier = Modifier.fillMaxSize()) {
            CompositionLocalProvider(LocalTvSharedTransitionScope provides this@SharedTransitionLayout) {
                NavHost(
                    navController = navController,
                    startDestination = "tv-home",
                    modifier = Modifier
                        .fillMaxSize()
                        .background(AppChrome.PageGradient),
                ) {
                    composable("tv-home") {
                        com.chee.videos.feature.tv.TvCatalogScreen(
                            onOpenSeries = { seriesId ->
                                navController.navigate(buildTvSeriesRoute(seriesId))
                            },
                            onOpenContinueWatching = { seriesId, season, episode ->
                                navController.navigate(buildTvPlayerRoute(seriesId, season, episode))
                            },
                            onOpenLongForm = { videoId, videoType ->
                                navController.navigate(buildTvLongFormDetailRoute(videoId, videoType))
                            },
                            onPlayLongForm = { videoId, videoType ->
                                navController.navigate(buildTvLongFormPlayerRoute(videoId, videoType))
                            },
                            onOpenCatalogWall = { kind, title ->
                                navController.navigate(buildTvCatalogWallRoute(kind, title))
                            },
                            onOpenIptv = {
                                navController.navigate(TvIptvRoute)
                            },
                            homeContentFocusRequester = homeContentFocusRequester,
                            onRepair = onRepair,
                            onLogout = onLogout,
                            onSwitchServer = onSwitchServer,
                        )
                    }
                    composable(
                        route = TvCatalogWallRoutePattern,
                        arguments = listOf(
                            navArgument(TvCatalogWallKindArg) { type = NavType.StringType },
                            navArgument(TvCatalogWallTitleArg) {
                                type = NavType.StringType
                                defaultValue = ""
                            },
                        ),
                    ) {
                        CompositionLocalProvider(LocalTvAnimatedContentScope provides this@composable) {
                            com.chee.videos.feature.tv.TvPosterWallScreen(
                                baseUrl = baseUrl,
                                onBack = { navController.popBackStack() },
                                onOpenSeries = { seriesId -> navController.navigate(buildTvSeriesRoute(seriesId)) },
                                onOpenLongForm = { videoId, videoType ->
                                    navController.navigate(buildTvLongFormDetailRoute(videoId, videoType))
                                },
                            )
                        }
                    }
                    composable(
                        route = TvLongFormDetailRoutePattern,
                        arguments = listOf(
                            navArgument(TvLongFormVideoIdArg) { type = NavType.StringType },
                            navArgument(TvLongFormVideoTypeArg) {
                                type = NavType.StringType
                                defaultValue = "movie"
                            },
                        ),
                    ) {
                        TvLongFormDetailScreen(
                            onBack = { navController.popBackStack() },
                            onPlay = { videoId, videoType ->
                                navController.navigate(buildTvLongFormPlayerRoute(videoId, videoType))
                            },
                        )
                    }
                    composable(
                        route = TvLongFormPlayerRoutePattern,
                        arguments = listOf(
                            navArgument(TvLongFormVideoIdArg) { type = NavType.StringType },
                            navArgument(TvLongFormVideoTypeArg) {
                                type = NavType.StringType
                                defaultValue = "movie"
                            },
                        ),
                    ) {
                        TvLongFormPlayerScreen(
                            onBack = { navController.popBackStack() },
                        )
                    }
                    composable(TvIptvRoute) {
                        TvIptvScreen(
                            onBack = { navController.popBackStack() },
                        )
                    }
                    composable(
                        route = TvSeriesRoutePattern,
                        arguments = listOf(navArgument(TvSeriesIdArg) { type = NavType.StringType }),
                    ) {
                        CompositionLocalProvider(LocalTvAnimatedContentScope provides this@composable) {
                            TvSeriesDetailScreen(
                                onBack = { navController.popBackStack() },
                                onPlayEpisode = { seriesId, season, episode ->
                                    navController.navigate(buildTvPlayerRoute(seriesId, season, episode))
                                },
                            )
                        }
                    }
                    composable(
                        route = TvPlayerRoutePattern,
                        arguments = listOf(
                            navArgument(TvSeriesIdArg) { type = NavType.StringType },
                            navArgument(TvSeasonArg) { type = NavType.IntType; defaultValue = 1 },
                            navArgument(TvEpisodeArg) { type = NavType.IntType; defaultValue = 1 },
                        ),
                    ) {
                        TvSeriesPlayerScreen(
                            accessToken = accessToken,
                            onBack = { navController.popBackStack() },
                        )
                    }
                }
            }
        }
        }

        if (showRootExitPrompt) {
            TvRootExitConfirmPrompt(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 48.dp),
            )
        }
    }
}

internal fun shouldShowTvGlobalSettings(route: String?): Boolean = route == "__disabled_tv_global_settings__"

internal const val TvRootExitConfirmWindowMillis = 2_000L

internal enum class TvRootBackAction {
    ShowPrompt,
    Exit,
}

internal fun shouldHandleTvRootExitConfirm(route: String?): Boolean = route == "tv-home"

internal fun resolveTvRootBackAction(
    previousPromptUptimeMillis: Long?,
    nowUptimeMillis: Long,
    confirmWindowMillis: Long = TvRootExitConfirmWindowMillis,
): TvRootBackAction {
    val previous = previousPromptUptimeMillis ?: return TvRootBackAction.ShowPrompt
    val elapsed = nowUptimeMillis - previous
    return if (elapsed in 0..confirmWindowMillis) {
        TvRootBackAction.Exit
    } else {
        TvRootBackAction.ShowPrompt
    }
}

@Composable
private fun TvRootExitConfirmPrompt(
    modifier: Modifier = Modifier,
) {
    Surface(
        modifier = modifier,
        color = Color(0xCC121212),
        shape = RoundedCornerShape(14.dp),
    ) {
        Text(
            text = "再按一次退出",
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
            color = Color.White,
            style = MaterialTheme.typography.bodySmall,
        )
    }
}

private tailrec fun Context.findActivity(): Activity? = when (this) {
    is Activity -> this
    is ContextWrapper -> baseContext.findActivity()
    else -> null
}

internal enum class TvShellSettingsFocusDirection {
    Left,
    Down,
    Right,
    Up,
}

internal enum class TvShellSettingsFocusTarget {
    HomeContent,
    Boundary,
}

internal fun resolveTvShellSettingsFocusTarget(
    direction: TvShellSettingsFocusDirection,
): TvShellSettingsFocusTarget = when (direction) {
    TvShellSettingsFocusDirection.Left,
    TvShellSettingsFocusDirection.Down,
    -> TvShellSettingsFocusTarget.HomeContent

    TvShellSettingsFocusDirection.Right,
    TvShellSettingsFocusDirection.Up,
    -> TvShellSettingsFocusTarget.Boundary
}

@OptIn(ExperimentalComposeUiApi::class)
private fun tvShellSettingsFocusRequesterFor(
    direction: TvShellSettingsFocusDirection,
    homeContentFocusRequester: FocusRequester,
): FocusRequester = when (resolveTvShellSettingsFocusTarget(direction)) {
    TvShellSettingsFocusTarget.HomeContent -> homeContentFocusRequester
    TvShellSettingsFocusTarget.Boundary -> FocusRequester.Cancel
}

internal fun shouldHandleTvShellBack(route: String?): Boolean {
    val value = route ?: return false
    if (value == "tv-home") {
        return false
    }
    if (value.isTvPlaybackRoute()) {
        return false
    }
    return value == TvCatalogWallRoutePattern ||
        value.startsWith("tv/wall/") ||
        value == TvLongFormDetailRoutePattern ||
        value.startsWith("tv/detail/") ||
        value == TvIptvRoute ||
        value == TvSeriesRoutePattern ||
        value.startsWith("tv/series/")
}

private fun String.isTvPlaybackRoute(): Boolean =
    this == TvPlayerRoutePattern ||
        startsWith("tv/player/") ||
        this == TvLongFormPlayerRoutePattern ||
        startsWith("tv/long-form-player/") ||
        this == TvIptvRoute
