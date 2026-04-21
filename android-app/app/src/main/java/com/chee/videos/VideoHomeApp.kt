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
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
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
import com.chee.videos.core.model.AppRootState
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.AppDarkColors
import com.chee.videos.core.ui.rootNavigationTabs
import com.chee.videos.core.viewmodel.AppRootViewModel
import com.chee.videos.feature.auth.LoginScreen
import com.chee.videos.feature.connection.ConnectionScreen
import com.chee.videos.feature.detail.DetailScreen
import com.chee.videos.feature.home.HomeScreen
import com.chee.videos.feature.mine.MineScreen
import com.chee.videos.feature.player.UnifiedPlayerScreen
import com.chee.videos.feature.shortdiscover.ShortDiscoverScreen

@Composable
fun VideoHomeApp(
    appRootViewModel: AppRootViewModel = hiltViewModel(),
) {
    val appState by appRootViewModel.appState.collectAsStateWithLifecycle()

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
                    AuthenticatedNav(
                        baseUrl = state.baseUrl,
                        accessToken = state.accessToken,
                        onSwitchServer = appRootViewModel::switchToServerSelection,
                        onLogout = appRootViewModel::logout,
                    )
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
) {
    val navController = rememberNavController()
    val backStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = backStackEntry?.destination?.route.orEmpty()
    val showBottomBar = currentRoute == "home" || currentRoute == "mine"

    Scaffold(
        containerColor = AppChrome.Canvas,
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
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
                        onOpenShortDiscover = { mode, value, title ->
                            navController.navigate(
                                "short-discover/${Uri.encode(mode)}/${Uri.encode(value)}/${Uri.encode(title)}",
                            )
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
                        onSwitchServer = onSwitchServer,
                        onLogout = onLogout,
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
        }
    }
}
