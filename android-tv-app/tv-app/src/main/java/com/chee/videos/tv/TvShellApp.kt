package com.chee.videos.tv

import androidx.compose.foundation.background
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.chee.videos.core.model.AppRootState
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.AppDarkColors
import com.chee.videos.core.viewmodel.AppRootViewModel
import com.chee.videos.feature.connection.ConnectionScreen
import com.chee.videos.feature.detail.DetailScreen
import com.chee.videos.feature.tv.TvEpisodeArg
import com.chee.videos.feature.tv.TvPlayerRoutePattern
import com.chee.videos.feature.tv.TvSeasonArg
import com.chee.videos.feature.tv.TvSeriesDetailScreen
import com.chee.videos.feature.tv.TvSeriesIdArg
import com.chee.videos.feature.tv.TvSeriesPlayerScreen
import com.chee.videos.feature.tv.TvSeriesRoutePattern
import com.chee.videos.feature.tv.buildTvPlayerRoute

@Composable
fun TvShellApp(
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

                AppRootState.NeedServer -> ConnectionScreen()
                AppRootState.NeedLogin -> TvPairingScreen(
                    onSwitchServer = appRootViewModel::switchToServerSelection,
                )
                is AppRootState.Ready -> TvAuthenticatedNav(
                    accessToken = state.accessToken,
                    onLogout = appRootViewModel::logout,
                    onRepair = appRootViewModel::logout,
                    onSwitchServer = appRootViewModel::switchToServerSelection,
                )
            }
        }
    }
}

@Composable
private fun TvAuthenticatedNav(
    accessToken: String,
    onLogout: () -> Unit,
    onRepair: () -> Unit,
    onSwitchServer: () -> Unit,
) {
    val navController = rememberNavController()
    var menuExpanded by remember { mutableStateOf(false) }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient),
    ) {
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
                        navController.navigate("tv/series/$seriesId")
                    },
                    onOpenContinueWatching = { seriesId, season, episode ->
                        navController.navigate(buildTvPlayerRoute(seriesId, season, episode))
                    },
                    onOpenLongForm = { videoId, videoType ->
                        navController.navigate("detail/$videoId?type=$videoType")
                    },
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
                DetailScreen(onBack = { navController.popBackStack() })
            }
            composable(
                route = TvSeriesRoutePattern,
                arguments = listOf(navArgument(TvSeriesIdArg) { type = NavType.StringType }),
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
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .statusBarsPadding()
                .padding(horizontal = 18.dp, vertical = 12.dp),
        ) {
            Box(modifier = Modifier.weight(1f))
            Column(horizontalAlignment = Alignment.End) {
                IconButton(
                    onClick = { menuExpanded = true },
                    modifier = Modifier.focusable(),
                ) {
                    Icon(
                        imageVector = Icons.Filled.Settings,
                        contentDescription = "账户与设备菜单",
                        tint = AppChrome.TextPrimary,
                    )
                }
                DropdownMenu(
                    expanded = menuExpanded,
                    onDismissRequest = { menuExpanded = false },
                ) {
                    TvAccountMenuAction.defaults().forEach { action ->
                        DropdownMenuItem(
                            text = { Text(action.label, color = AppChrome.TextPrimary) },
                            onClick = {
                                menuExpanded = false
                                when (action) {
                                    TvAccountMenuAction.Repair -> onRepair()
                                    TvAccountMenuAction.Logout -> onLogout()
                                    TvAccountMenuAction.SwitchServer -> onSwitchServer()
                                }
                            },
                        )
                    }
                }
            }
        }
    }
}
