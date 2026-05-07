package com.chee.videos

import android.net.Uri
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
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.google.zxing.client.android.Intents
import com.chee.videos.core.model.AppRootState
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.AppDarkColors
import com.chee.videos.core.ui.rootNavigationTabs
import com.chee.videos.core.viewmodel.AppRootViewModel
import com.chee.videos.feature.auth.LoginScreen
import com.chee.videos.feature.connection.ConnectionScreen
import com.chee.videos.feature.detail.DetailScreen
import com.chee.videos.feature.home.HomeScreen
import com.chee.videos.feature.imagecollections.ImageCollectionViewerScreen
import com.chee.videos.feature.imagecollections.ImageCollectionsScreen
import com.chee.videos.feature.mine.MineScreen
import com.chee.videos.feature.player.UnifiedPlayerScreen
import com.chee.videos.feature.shortdiscover.ShortDiscoverScreen
import com.chee.videos.feature.shortsearch.ShortSearchScreen
import com.chee.videos.feature.tv.TvEpisodeArg
import com.chee.videos.feature.tv.TvPlayerRoutePattern
import com.chee.videos.feature.tv.TvSeasonArg
import com.chee.videos.feature.tv.TvSeriesDetailScreen
import com.chee.videos.feature.tv.TvSeriesIdArg
import com.chee.videos.feature.tv.TvSeriesPlayerScreen
import com.chee.videos.feature.tv.TvSeriesRoutePattern
import com.chee.videos.feature.tvauth.TvAuthApprovalScreen
import com.chee.videos.feature.tvauth.TvAuthDeepLinkParser
import com.chee.videos.feature.tvauth.resolveTvAuthDeepLink
import com.chee.videos.feature.tv.buildTvPlayerRoute
import com.chee.videos.feature.tv.buildTvSeriesRoute
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

    Scaffold(
        containerColor = AppChrome.Canvas,
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
        snackbarHost = {
            SnackbarHost(hostState = snackbarHostState)
        },
        bottomBar = {
            if (showBottomBar) {
                Surface(
                    color = AppChrome.Surface,
                    contentColor = AppChrome.TextPrimary,
                    shadowElevation = 18.dp,
                    shape = RoundedCornerShape(topStart = 26.dp, topEnd = 26.dp),
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
                                        .clip(RoundedCornerShape(16.dp))
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
                            navController.navigate("detail/$videoId?type=${Uri.encode(videoType)}")
                        },
                        onOpenTvSeries = { seriesId ->
                            navController.navigate(buildTvSeriesRoute(seriesId))
                        },
                        onOpenTvContinueWatching = { seriesId, season, episode ->
                            navController.navigate(buildTvPlayerRoute(seriesId, season, episode))
                        },
                        onOpenShortDiscover = { mode, value, title ->
                            navController.navigate(
                                "short-discover/${Uri.encode(mode)}/${Uri.encode(value)}/${Uri.encode(title)}",
                            )
                        },
                        onOpenImageCollectionViewer = { route ->
                            navController.navigate(route)
                        },
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
                                    setOrientationLocked(false)
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
                    )
                }
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
                DetailScreen(onBack = { navController.popBackStack() })
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
                route = TvSeriesRoutePattern,
                arguments = listOf(
                    navArgument(TvSeriesIdArg) { type = NavType.StringType },
                ),
            ) {
                TvSeriesDetailScreen(
                    onBack = { navController.popBackStack() },
                    onPlayEpisode = { seriesId, season, episode ->
                        navController.navigate(buildTvPlayerRoute(seriesId, season, episode))
                    },
                )
            }

            composable(
                route = TvPlayerRoutePattern,
                arguments = listOf(
                    navArgument(TvSeriesIdArg) { type = NavType.StringType },
                    navArgument(TvSeasonArg) {
                        type = NavType.IntType
                        defaultValue = 1
                    },
                    navArgument(TvEpisodeArg) {
                        type = NavType.IntType
                        defaultValue = 1
                    },
                ),
            ) {
                TvSeriesPlayerScreen(
                    accessToken = accessToken,
                    onBack = { navController.popBackStack() },
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
