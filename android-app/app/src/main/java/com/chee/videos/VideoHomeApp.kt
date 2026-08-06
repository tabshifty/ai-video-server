package com.chee.videos

import android.net.Uri
import androidx.compose.animation.AnimatedContentTransitionScope
import androidx.compose.animation.EnterTransition
import androidx.compose.animation.ExitTransition
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavBackStackEntry
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.google.zxing.client.android.Intents
import com.chee.videos.core.model.AppRootState
import com.chee.videos.core.ui.AppNavigationTransitionDirection
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.AppDarkColors
import com.chee.videos.core.ui.appNavigationTransitionDirection
import com.chee.videos.core.ui.appNavigationTransitionSpec
import com.chee.videos.core.ui.rootNavigationTabs
import com.chee.videos.core.viewmodel.AppRootViewModel
import com.chee.videos.feature.actor.ActorDetailScreen
import com.chee.videos.feature.actor.ActorIdArg
import com.chee.videos.feature.actor.ActorRoutePattern
import com.chee.videos.feature.actor.buildActorRoute
import com.chee.videos.feature.actor.buildVideoDetailRoute
import com.chee.videos.feature.auth.LoginScreen
import com.chee.videos.feature.connection.ConnectionScreen
import com.chee.videos.feature.detail.DetailScreen
import com.chee.videos.feature.home.HomeScreen
import com.chee.videos.feature.imagecollections.ImageCollectionViewerScreen
import com.chee.videos.feature.imagecollections.ImageCollectionsScreen
import com.chee.videos.feature.mine.MineScreen
import com.chee.videos.feature.player.UnifiedPlayerScreen
import com.chee.videos.feature.shortdiscover.ShortDiscoverScreen
import com.chee.videos.feature.shortcollections.ShortCollectionContentRoutePattern
import com.chee.videos.feature.shortcollections.ShortCollectionContentScreen
import com.chee.videos.feature.shortcollections.ShortCollectionIdArg
import com.chee.videos.feature.shortcollections.ShortCollectionNameArg
import com.chee.videos.feature.shortcollections.ShortCollectionsViewModel
import com.chee.videos.feature.shortcollections.buildShortCollectionContentRoute
import com.chee.videos.feature.shortsearch.ShortSearchScreen
import com.chee.videos.feature.shortsearch.ShortSearchRemoteControlRoutePattern
import com.chee.videos.feature.shortsearch.ShortSearchRemoteControlScreen
import com.chee.videos.feature.shortsearch.buildShortSearchRemoteControlRoute
import com.chee.videos.feature.tvauth.TvAuthApprovalScreen
import com.chee.videos.feature.tvauth.TvAuthDeepLinkParser
import com.chee.videos.feature.tvauth.PortraitCaptureActivity
import com.chee.videos.feature.tvauth.resolveTvAuthDeepLink
import com.journeyapps.barcodescanner.ScanContract
import com.journeyapps.barcodescanner.ScanOptions
import kotlinx.coroutines.launch

@Composable
fun VideoHomeApp(
    appRootViewModel: AppRootViewModel = hiltViewModel(),
    initialTvAuthDeepLink: String? = null,
    onConsumeTvAuthDeepLink: () -> Unit = {},
) {
    val appState by appRootViewModel.appState.collectAsStateWithLifecycle()
    var scannedTvAuthPayload by rememberSaveable { mutableStateOf<String?>(null) }
    val tvAuthDeepLink = resolveTvAuthDeepLink(
        launchPayload = initialTvAuthDeepLink,
        scannedPayload = scannedTvAuthPayload,
    )

    LaunchedEffect(tvAuthDeepLink?.serverBaseUrl) {
        appRootViewModel.applyTvAuthServer(tvAuthDeepLink?.serverBaseUrl)
    }

    MaterialTheme(colorScheme = AppDarkColors) {
        Surface(modifier = Modifier.fillMaxSize(), color = AppChrome.Canvas) {
            when (val state = appState) {
                AppRootState.Loading -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = AppChrome.AccentStrong)
                    }
                }

                AppRootState.NeedServer -> {
                    ConnectionScreen()
                }

                AppRootState.NeedLogin -> {
                    LoginScreen(onSwitchServer = { appRootViewModel.switchToServerSelection() })
                }

                is AppRootState.Ready -> {
                    if (tvAuthDeepLink != null) {
                        TvAuthApprovalScreen(
                            deepLink = tvAuthDeepLink,
                            onFinished = {
                                scannedTvAuthPayload = null
                                onConsumeTvAuthDeepLink()
                            },
                        )
                    } else {
                        AuthenticatedNav(
                            baseUrl = state.baseUrl,
                            accessToken = state.accessToken,
                            onSwitchServer = appRootViewModel::switchToServerSelection,
                            onLogout = appRootViewModel::logout,
                            onScannedTvAuthPayload = { payload ->
                                scannedTvAuthPayload = payload
                            },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun AuthenticatedNav(
    baseUrl: String,
    accessToken: String,
    onSwitchServer: () -> Unit,
    onLogout: () -> Unit,
    onScannedTvAuthPayload: (String) -> Unit,
) {
    val navController = rememberNavController()
    val backStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = backStackEntry?.destination?.route.orEmpty()
    val showBottomBar = rootNavigationTabs.any { it.route == currentRoute }
    val snackbarHostState = remember { SnackbarHostState() }
    val scope = rememberCoroutineScope()
    var isShortFullscreen by rememberSaveable { mutableStateOf(false) }
    val scannerLauncher = androidx.activity.compose.rememberLauncherForActivityResult(ScanContract()) { result ->
        val contents = result.contents?.trim().orEmpty()
        when {
            contents.isBlank() -> Unit
            TvAuthDeepLinkParser.parse(contents) != null -> onScannedTvAuthPayload(contents)
            else -> {
                scope.launch {
                    snackbarHostState.showSnackbar("未识别为 TV 登录二维码")
                }
            }
        }
    }
    val returnFromUnavailableContent: () -> Unit = {
        scope.launch {
            snackbarHostState.showSnackbar("该内容不在手机端提供")
        }
        if (!navController.popBackStack()) {
            navController.navigate("home") {
                popUpTo(navController.graph.findStartDestination().id)
                launchSingleTop = true
            }
        }
    }

    Scaffold(
        containerColor = AppChrome.Canvas,
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
        snackbarHost = {
            SnackbarHost(hostState = snackbarHostState)
        },
        bottomBar = {
            if (showBottomBar && !isShortFullscreen) {
                Surface(
                    color = AppChrome.Surface,
                    contentColor = AppChrome.TextPrimary,
                    shadowElevation = 18.dp,
                    shape = RoundedCornerShape(topStart = 12.dp, topEnd = 12.dp),
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(AppChrome.Surface),
                    ) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(1.dp)
                                .background(AppChrome.Divider),
                        )
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .navigationBarsPadding()
                                .padding(horizontal = 12.dp, vertical = 8.dp),
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            rootNavigationTabs.forEach { tab ->
                                val selected = currentRoute == tab.route
                                Box(
                                    modifier = Modifier
                                        .weight(1f)
                                        .clip(RoundedCornerShape(8.dp))
                                        .clickable {
                                            navController.navigate(tab.route) {
                                                popUpTo(navController.graph.findStartDestination().id) {
                                                    saveState = true
                                                }
                                                launchSingleTop = true
                                                restoreState = true
                                            }
                                        }
                                        .background(if (selected) AppChrome.AccentSoft else Color.Transparent)
                                        .padding(vertical = 12.dp),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    Text(
                                        text = tab.label,
                                        style = MaterialTheme.typography.titleSmall,
                                        fontWeight = FontWeight.SemiBold,
                                        color = if (selected) AppChrome.TextPrimary else AppChrome.TextMuted,
                                    )
                                }
                            }
                        }
                    }
                }
            }
        },
    ) { innerPadding ->
        NavHost(
            navController = navController,
            startDestination = "home",
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding),
            enterTransition = {
                appNavigationEnterTransition(isPop = false)
            },
            exitTransition = {
                appNavigationExitTransition(isPop = false)
            },
            popEnterTransition = {
                appNavigationEnterTransition(isPop = true)
            },
            popExitTransition = {
                appNavigationExitTransition(isPop = true)
            },
        ) {
            composable("home") {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(AppChrome.PageGradient),
                ) {
                    HomeScreen(
                        baseUrl = baseUrl,
                        accessToken = accessToken,
                        onOpenDetail = { videoId, videoType ->
                            navController.navigate(buildVideoDetailRoute(videoId, videoType))
                        },
                        onOpenShortDiscover = { mode, value, title ->
                            navController.navigate(
                                "short-discover/${Uri.encode(mode)}/${Uri.encode(value)}/${Uri.encode(title)}",
                            )
                        },
                        onOpenShortCollection = { collectionId, collectionName ->
                            navController.navigate(buildShortCollectionContentRoute(collectionId, collectionName))
                        },
                        onOpenImageCollectionViewer = { route ->
                            navController.navigate(route)
                        },
                        onShortFullscreenChange = { isShortFullscreen = it },
                    )
                }
            }

            composable("mine") {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(AppChrome.PageGradient),
                ) {
                    MineScreen(
                        baseUrl = baseUrl,
                        onOpenPlayer = { source, videoId ->
                            navController.navigate("player/$source/$videoId")
                        },
                        onScanTvLogin = {
                            scannerLauncher.launch(
                                ScanOptions().apply {
                                    setDesiredBarcodeFormats(ScanOptions.QR_CODE)
                                    setPrompt("扫描 TV 登录二维码")
                                    setBeepEnabled(false)
                                    // 用自定义竖屏 CaptureActivity 覆盖库默认的横屏 Activity；
                                    // 方向由 Manifest 锁定，这里不再用 setOrientationLocked。
                                    setCaptureActivity(PortraitCaptureActivity::class.java)
                                    addExtra(Intents.Scan.FORMATS, ScanOptions.QR_CODE)
                                },
                            )
                        },
                        onSwitchServer = onSwitchServer,
                        onLogout = onLogout,
                    )
                }
            }

            composable("search") {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(AppChrome.PageGradient),
                ) {
                    ShortSearchScreen(
                        baseUrl = baseUrl,
                        accessToken = accessToken,
                        onFullscreenChange = { isShortFullscreen = it },
                        onOpenRemoteControl = { sessionId ->
                            navController.navigate(buildShortSearchRemoteControlRoute(sessionId))
                        },
                    )
                }
            }

            composable(
                route = ShortSearchRemoteControlRoutePattern,
                arguments = listOf(navArgument("sessionId") { type = NavType.StringType }),
            ) {
                ShortSearchRemoteControlScreen(
                    onBack = { navController.popBackStack() },
                )
            }

            composable(
                route = "detail/{videoId}?type={videoType}",
                arguments = listOf(
                    navArgument("videoId") { type = NavType.StringType },
                    navArgument("videoType") {
                        type = NavType.StringType
                        defaultValue = ""
                    },
                ),
            ) {
                DetailScreen(
                    onBack = { navController.popBackStack() },
                    onUnsupportedContent = returnFromUnavailableContent,
                    onOpenActor = { actorId ->
                        navController.navigate(buildActorRoute(actorId))
                    },
                )
            }

            composable(
                route = ActorRoutePattern,
                arguments = listOf(
                    navArgument(ActorIdArg) { type = NavType.StringType },
                ),
            ) { entry ->
                ActorDetailScreen(
                    actorId = entry.arguments?.getString(ActorIdArg).orEmpty(),
                    baseUrl = baseUrl,
                    onBack = { navController.popBackStack() },
                    onOpenDetail = { videoId, videoType ->
                        navController.navigate(buildVideoDetailRoute(videoId, videoType))
                    },
                )
            }

            composable(
                route = "player/{source}/{videoId}",
                arguments = listOf(
                    navArgument("source") { type = NavType.StringType },
                    navArgument("videoId") { type = NavType.StringType },
                ),
            ) { entry ->
                UnifiedPlayerScreen(
                    baseUrl = baseUrl,
                    accessToken = accessToken,
                    source = entry.arguments?.getString("source").orEmpty(),
                    startVideoId = entry.arguments?.getString("videoId").orEmpty(),
                    onBack = { navController.popBackStack() },
                    onUnsupportedContent = returnFromUnavailableContent,
                )
            }

            composable(
                route = "short-discover/{mode}/{value}/{title}",
                arguments = listOf(
                    navArgument("mode") { type = NavType.StringType },
                    navArgument("value") { type = NavType.StringType },
                    navArgument("title") { type = NavType.StringType },
                ),
            ) { entry ->
                ShortDiscoverScreen(
                    baseUrl = baseUrl,
                    accessToken = accessToken,
                    mode = entry.arguments?.getString("mode").orEmpty(),
                    value = entry.arguments?.getString("value").orEmpty(),
                    title = entry.arguments?.getString("title").orEmpty(),
                    onBack = { navController.popBackStack() },
                )
            }

            composable(
                route = ShortCollectionContentRoutePattern,
                arguments = listOf(
                    navArgument(ShortCollectionIdArg) { type = NavType.StringType },
                    navArgument(ShortCollectionNameArg) {
                        type = NavType.StringType
                        defaultValue = "短视频合集"
                    },
                ),
            ) { entry ->
                val homeEntry = remember(entry) { navController.getBackStackEntry("home") }
                val directoryViewModel: ShortCollectionsViewModel = hiltViewModel(homeEntry)
                val collectionUnavailable = {
                    directoryViewModel.refresh()
                    navController.popBackStack()
                    Unit
                }
                ShortCollectionContentScreen(
                    baseUrl = baseUrl,
                    accessToken = accessToken,
                    collectionId = entry.arguments?.getString(ShortCollectionIdArg).orEmpty(),
                    collectionName = entry.arguments?.getString(ShortCollectionNameArg).orEmpty(),
                    onBack = { navController.popBackStack() },
                    onCollectionUnavailable = collectionUnavailable,
                    onFullscreenChange = { isShortFullscreen = it },
                    onOpenRemoteControl = { sessionId ->
                        navController.navigate(buildShortSearchRemoteControlRoute(sessionId))
                    },
                )
            }

            composable("image-collections") {
                ImageCollectionsScreen(
                    baseUrl = baseUrl,
                    onBack = null,
                    onOpenCollection = { collectionId ->
                        navController.navigate("image-collections/$collectionId")
                    },
                )
            }

            composable(
                route = "image-collections/{collectionId}",
                arguments = listOf(
                    navArgument("collectionId") { type = NavType.StringType },
                ),
            ) {
                ImageCollectionViewerScreen(
                    baseUrl = baseUrl,
                    onBack = { navController.popBackStack() },
                )
            }
        }
    }
}

private fun AnimatedContentTransitionScope<NavBackStackEntry>.appNavigationEnterTransition(
    isPop: Boolean,
): EnterTransition {
    val spec = appNavigationTransitionSpec()
    val direction = appNavigationTransitionDirection(
        fromRoute = initialState.destination.route,
        toRoute = targetState.destination.route,
        isPop = isPop,
    )
    val slideDirection = when (direction) {
        AppNavigationTransitionDirection.Forward -> AnimatedContentTransitionScope.SlideDirection.Left
        AppNavigationTransitionDirection.Backward -> AnimatedContentTransitionScope.SlideDirection.Right
    }
    return slideIntoContainer(slideDirection, animationSpec = tween(spec.durationMillis)) +
        fadeIn(animationSpec = tween(spec.durationMillis), initialAlpha = spec.fadeStartAlpha)
}

private fun AnimatedContentTransitionScope<NavBackStackEntry>.appNavigationExitTransition(
    isPop: Boolean,
): ExitTransition {
    val spec = appNavigationTransitionSpec()
    val direction = appNavigationTransitionDirection(
        fromRoute = initialState.destination.route,
        toRoute = targetState.destination.route,
        isPop = isPop,
    )
    val slideDirection = when (direction) {
        AppNavigationTransitionDirection.Forward -> AnimatedContentTransitionScope.SlideDirection.Left
        AppNavigationTransitionDirection.Backward -> AnimatedContentTransitionScope.SlideDirection.Right
    }
    return slideOutOfContainer(slideDirection, animationSpec = tween(spec.durationMillis)) +
        fadeOut(animationSpec = tween(spec.durationMillis))
}
