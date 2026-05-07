package com.chee.videos.tv

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
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
                AppRootState.NeedLogin -> TvPairingScreen()
                is AppRootState.Ready -> TvAuthenticatedNav(accessToken = state.accessToken)
            }
        }
    }
}

@Composable
private fun TvAuthenticatedNav(accessToken: String) {
    val navController = rememberNavController()

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
            )
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
}
