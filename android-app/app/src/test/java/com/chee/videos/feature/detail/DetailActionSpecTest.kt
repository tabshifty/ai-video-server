package com.chee.videos.feature.detail

import com.chee.videos.core.model.UserStateDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class DetailActionSpecTest {

    @Test
    fun `uses two primary actions and one secondary danger action`() {
        val specs = buildDetailActionSpecs(UserStateDto())

        assertEquals(
            listOf(
                DetailActionSpec(
                    action = DetailAction.Like,
                    label = "点赞",
                    active = false,
                    tone = DetailActionTone.Primary,
                ),
                DetailActionSpec(
                    action = DetailAction.Favorite,
                    label = "收藏",
                    active = false,
                    tone = DetailActionTone.Primary,
                ),
                DetailActionSpec(
                    action = DetailAction.Dislike,
                    label = "不喜欢",
                    active = false,
                    tone = DetailActionTone.SecondaryDanger,
                ),
            ),
            specs,
        )
    }

    @Test
    fun `reflects selected user state in action labels and active states`() {
        val specs = buildDetailActionSpecs(
            UserStateDto(
                isLiked = true,
                isFavorited = true,
                isDisliked = true,
            ),
        )

        assertEquals("取消点赞", specs[0].label)
        assertTrue(specs[0].active)
        assertEquals("取消收藏", specs[1].label)
        assertTrue(specs[1].active)
        assertEquals("取消不喜欢", specs[2].label)
        assertTrue(specs[2].active)
        assertFalse(specs[2].tone == DetailActionTone.Primary)
    }
}
