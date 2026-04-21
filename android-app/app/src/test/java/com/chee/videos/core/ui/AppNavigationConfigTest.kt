package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Test

class AppNavigationConfigTest {
    @Test
    fun homeContentTabs_matchCompactCategories() {
        assertEquals(
            listOf("短视频" to "short", "电影" to "movie", "电视剧" to "episode", "AV" to "av"),
            homeContentTabs.map { it.title to it.type },
        )
    }

    @Test
    fun homeContentTabs_doNotCarryExtraPresentationFields() {
        assertEquals(
            setOf("title", "type"),
            HomeContentTabSpec::class.java.declaredFields
                .filterNot { it.isSynthetic || it.name.startsWith("$") }
                .map { it.name }
                .toSet(),
        )
    }

    @Test
    fun rootNavigationTabs_matchCompactBottomNavigation() {
        assertEquals(
            listOf("home" to "首页", "mine" to "我的"),
            rootNavigationTabs.map { it.route to it.label },
        )
    }

    @Test
    fun rootNavigationTabs_doNotCarryExtraPresentationFields() {
        assertEquals(
            setOf("route", "label"),
            RootNavigationTabSpec::class.java.declaredFields
                .filterNot { it.isSynthetic || it.name.startsWith("$") }
                .map { it.name }
                .toSet(),
        )
    }
}
