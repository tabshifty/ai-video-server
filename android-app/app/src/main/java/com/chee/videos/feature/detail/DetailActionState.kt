package com.chee.videos.feature.detail

import com.chee.videos.core.model.UserStateDto

internal enum class DetailAction {
    Like,
    Favorite,
    Dislike,
}

internal enum class DetailActionTone {
    Primary,
    SecondaryDanger,
}

internal data class DetailActionSpec(
    val action: DetailAction,
    val label: String,
    val active: Boolean,
    val tone: DetailActionTone,
)

internal fun buildDetailActionSpecs(userState: UserStateDto): List<DetailActionSpec> {
    return listOf(
        DetailActionSpec(
            action = DetailAction.Like,
            label = if (userState.isLiked) "取消点赞" else "点赞",
            active = userState.isLiked,
            tone = DetailActionTone.Primary,
        ),
        DetailActionSpec(
            action = DetailAction.Favorite,
            label = if (userState.isFavorited) "取消收藏" else "收藏",
            active = userState.isFavorited,
            tone = DetailActionTone.Primary,
        ),
        DetailActionSpec(
            action = DetailAction.Dislike,
            label = if (userState.isDisliked) "取消不喜欢" else "不喜欢",
            active = userState.isDisliked,
            tone = DetailActionTone.SecondaryDanger,
        ),
    )
}
