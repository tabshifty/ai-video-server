package com.chee.videos.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvAccountMenuActionTest {

    @Test
    fun defaults_exposeLogoutRepairAndServerSwitchInOrder() {
        assertEquals(
            listOf(
                TvAccountMenuAction.Repair,
                TvAccountMenuAction.Logout,
                TvAccountMenuAction.SwitchServer,
            ),
            TvAccountMenuAction.defaults(),
        )
    }
}
