package com.chee.videos.feature.tv

import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp

internal const val TvPlayerBackConfirmWindowMillis = 2_000L

internal enum class TvPlayerBackAction {
    ShowPrompt,
    Exit,
}

internal fun resolveTvPlayerBackAction(
    previousPromptUptimeMillis: Long?,
    nowUptimeMillis: Long,
    confirmWindowMillis: Long = TvPlayerBackConfirmWindowMillis,
): TvPlayerBackAction {
    val previous = previousPromptUptimeMillis ?: return TvPlayerBackAction.ShowPrompt
    val elapsed = nowUptimeMillis - previous
    return if (elapsed in 0..confirmWindowMillis) {
        TvPlayerBackAction.Exit
    } else {
        TvPlayerBackAction.ShowPrompt
    }
}

@Composable
internal fun TvPlayerBackConfirmPrompt(
    modifier: Modifier = Modifier,
) {
    Surface(
        modifier = modifier,
        color = Color(0xCC121212),
        shape = RoundedCornerShape(14.dp),
    ) {
        Text(
            text = "再按一次返回",
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
            color = Color.White,
            style = MaterialTheme.typography.bodySmall,
        )
    }
}
