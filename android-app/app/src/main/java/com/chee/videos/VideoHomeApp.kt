package com.chee.videos

import android.net.Uri
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
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
import com.chee.videos.core.viewmodel.AppRootViewModel
import com.chee.videos.feature.auth.LoginScreen
import com.chee.videos.feature.connection.ConnectionScreen
import com.chee.videos.feature.detail.DetailScreen
import com.chee.videos.feature.home.HomeScreen
import com.chee.videos.feature.mine.MineScreen
import com.chee.videos.feature.player.UnifiedPlayerScreen

private val AppDarkColors = darkColorScheme(
    primary = Color(0xFFFF5A7A),
    secondary = Color(0xFFFF5A7A),
    background = Color(0xFF090A0D),
    surface = Color(0xFF101318),
    onPrimary = Color.White,
    onSecondary = Color.White,
    onBackground = Color.White,
    onSurface = Color.White,
)

private data class RootTab(
    val route: String,
    val icon: androidx.compose.ui.graphics.vector.ImageVector,
    val label: String,
)

private val rootTabs = listOf(
    RootTab(route = "home", icon = Icons.Filled.Home, label = "首页"),
    RootTab(route = "mine", icon = Icons.Filled.Person, label = "我的"),
)

@Composable
fun VideoHomeApp(
    appRootViewModel: AppRootViewModel = hiltViewModel(),
) {
    val appState by appRootViewModel.appState.collectAsStateWithLifecycle()

    MaterialTheme(colorScheme = AppDarkColors) {
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
    val backStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = backStackEntry?.destination?.route.orEmpty()
    val showBottomBar = currentRoute == "home" || currentRoute == "mine"

    Scaffold(
        containerColor = Color(0xFF090A0D),
        bottomBar = {
            if (showBottomBar) {
                NavigationBar(containerColor = Color(0xFF101318), contentColor = Color.White) {
                    rootTabs.forEach { tab ->
                        val selected = currentRoute == tab.route
                        NavigationBarItem(
                            selected = selected,
                            onClick = {
                                navController.navigate(tab.route) {
                                    popUpTo(navController.graph.findStartDestination().id) {
                                        saveState = true
                                    }
                                    launchSingleTop = true
                                    restoreState = true
                                }
                            },
                            icon = { Icon(tab.icon, contentDescription = tab.label) },
                        )
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
                Box(modifier = Modifier.fillMaxSize()) {
                    HomeScreen(
                        baseUrl = baseUrl,
                        accessToken = accessToken,
                        onOpenDetail = { videoId, videoType ->
                            navController.navigate("detail/$videoId?type=${Uri.encode(videoType)}")
                        },
                    )
                }
            }

            composable("mine") {
                Box(modifier = Modifier.fillMaxSize()) {
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
        }
    }
}
