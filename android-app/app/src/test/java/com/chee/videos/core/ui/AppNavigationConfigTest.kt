package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Test

class AppNavigationConfigTest {
    @Test
    fun homeContentTabs_matchCompactCategories() {
        assertEquals(
            listOf(
                "短视频" to "short",
                "合集" to "short_collection",
                "AV" to "av",
            ),
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
    fun invalidRestoredHomeTabIndexFallsBackToShortVideo() {
        assertEquals(0, resolveHomeContentTabIndex(restoredIndex = -1, tabCount = 3))
        assertEquals(0, resolveHomeContentTabIndex(restoredIndex = 3, tabCount = 3))
        assertEquals(0, resolveHomeContentTabIndex(restoredIndex = 4, tabCount = 3))
        assertEquals(2, resolveHomeContentTabIndex(restoredIndex = 2, tabCount = 3))
    }

    @Test
    fun rootNavigationTabs_matchCompactBottomNavigation() {
        assertEquals(
            listOf("home" to "首页", "search" to "搜索", "image-collections" to "图集", "mine" to "我的"),
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
