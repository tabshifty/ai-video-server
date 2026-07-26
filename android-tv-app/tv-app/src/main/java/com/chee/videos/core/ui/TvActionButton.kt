package com.chee.videos.core.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp

/**
 * TV 端共享操作按钮（胶囊）。收敛各屏重复实现的 canonical 取值：
 * Primary = 暖金实底 + 深色画布前景，Secondary = 暗玻璃底 + 主文字色；
 * 内边距 18/12（满足 10-foot 焦点目标高度），titleSmall SemiBold；
 * 焦点缩放 Primary 1.06f / Secondary 1.05f（CONTEXT.md「TV 焦点反馈只缩放」的主操作分档）。
 * 例外：TvSeriesDetailScreen 按 CONTEXT.md「TV 参考图」授权保留本地实现，不接入本组件。
 */
enum class TvActionButtonTone { Primary, Secondary }

@Composable
fun TvActionButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    tone: TvActionButtonTone = TvActionButtonTone.Primary,
    icon: ImageVector? = null,
    enabled: Boolean = true,
    // 禁用时是否仍可聚焦：详情页「暂无片源」等作为首屏焦点目标的按钮必须保持可聚焦（默认），
    // 否则 10-foot 下整页可能无焦点落点；「扫描中」一类瞬时禁用按钮可传 false 跳过聚焦。
    focusableWhenDisabled: Boolean = true,
) {
    val containerColor = when {
        !enabled -> AppChrome.SurfaceStrong
        tone == TvActionButtonTone.Primary -> AppChrome.Accent
        else -> AppChrome.Surface.copy(alpha = 0.82f)
    }
    val contentColor = when {
        !enabled -> AppChrome.TextMuted
        tone == TvActionButtonTone.Primary -> AppChrome.Canvas
        else -> AppChrome.TextPrimary
    }
    val focusedScale = if (tone == TvActionButtonTone.Primary) 1.06f else 1.05f
    Surface(
        color = containerColor,
        shape = AppChrome.PillShape,
        modifier = modifier
            .tvFocusableScaleOnly(enabled = enabled || focusableWhenDisabled, focusedScale = focusedScale)
            .clickable(enabled = enabled, onClick = onClick),
    ) {
        // Surface 会向内容传播最小约束：外部传 fillMaxWidth 时 Row 被撑满，
        // CenterHorizontally 保证按钮文字居中；正常包裹内容时行为不变。
        Row(
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp, Alignment.CenterHorizontally),
        ) {
            if (icon != null) {
                Icon(icon, contentDescription = null, tint = contentColor)
            }
            Text(
                text = text,
                color = contentColor,
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}
