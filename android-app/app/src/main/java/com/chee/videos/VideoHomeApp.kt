package com.chee.videos

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
import com.chee.videos.core.viewmodel.AppRootViewModel
import com.chee.videos.feature.auth.LoginScreen
import com.chee.videos.feature.connection.ConnectionScreen
import com.chee.videos.feature.detail.DetailScreen
import com.chee.videos.feature.home.HomeScreen

@Composable
fun VideoHomeApp(
    appRootViewModel: AppRootViewModel = hiltViewModel(),
) {
    val appState by appRootViewModel.appState.collectAsStateWithLifecycle()

    MaterialTheme {
        Surface(modifier = Modifier.fillMaxSize()) {
            when (val state = appState) {
                AppRootState.Loading -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator()
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

    NavHost(
        navController = navController,
        startDestination = "home",
    ) {
        composable("home") {
            HomeScreen(
                baseUrl = baseUrl,
                accessToken = accessToken,
                onOpenDetail = { videoId -> navController.navigate("detail/$videoId") },
                onSwitchServer = onSwitchServer,
                onLogout = onLogout,
            )
        }
        composable(
            route = "detail/{videoId}",
            arguments = listOf(navArgument("videoId") { type = NavType.StringType }),
        ) {
            DetailScreen(onBack = { navController.popBackStack() })
        }
    }
}
