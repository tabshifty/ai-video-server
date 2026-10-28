package com.chee.videos.feature.tv

import androidx.compose.foundation.border
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.composed
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.drawOutline
import androidx.compose.ui.unit.dp
import com.chee.videos.core.ui.tvFocusableScaleOnly

// 设置页的局部例外：浅色边框表示焦点，暖金底色继续表示已选值。
internal fun Modifier.tvSettingsFocus(
    shape: Shape,
    enabled: Boolean = true,
    focusedScale: Float = 1.04f,
): Modifier = composed {
    var focused by remember { mutableStateOf(false) }
    val highlighted = enabled && focused
    this
        // 必须观察共享缩放修饰器创建的第一个焦点节点，也覆盖程序聚焦。
        .onFocusChanged { focused = it.hasFocus }
        .tvFocusableScaleOnly(enabled = enabled, focusedScale = focusedScale)
        .drawWithContent {
            drawContent()
            if (highlighted) {
                drawOutline(shape.createOutline(size, layoutDirection, this), Color.White.copy(alpha = 0.08f))
            }
        }
        .border(2.dp, if (highlighted) Color.White else Color.Transparent, shape)
}
