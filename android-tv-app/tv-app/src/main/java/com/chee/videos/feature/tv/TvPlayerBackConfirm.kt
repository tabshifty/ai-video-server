package com.chee.videos.feature.tv

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.chee.videos.core.ui.AppChrome

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
        color = AppChrome.SurfaceMuted.copy(alpha = 0.92f),
        shape = AppChrome.SurfaceShape,
    ) {
        Text(
            text = "再按一次返回",
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
            color = AppChrome.TextPrimary,
            style = MaterialTheme.typography.bodySmall,
        )
    }
}
