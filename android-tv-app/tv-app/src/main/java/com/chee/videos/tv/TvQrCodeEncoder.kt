package com.chee.videos.tv

import android.graphics.Bitmap
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.graphics.ImageBitmap
import com.google.zxing.BarcodeFormat
import com.google.zxing.EncodeHintType
import com.google.zxing.qrcode.QRCodeWriter

data class TvQrCodeMatrix(
    val width: Int,
    val height: Int,
    val rows: List<BooleanArray>,
)

object TvQrCodeEncoder {
    fun encode(content: String, size: Int): TvQrCodeMatrix {
        val bitMatrix = QRCodeWriter().encode(
            content,
            BarcodeFormat.QR_CODE,
            size,
            size,
            mapOf(EncodeHintType.MARGIN to 1),
        )
        return TvQrCodeMatrix(
            width = bitMatrix.width,
            height = bitMatrix.height,
            rows = List(bitMatrix.height) { y ->
                BooleanArray(bitMatrix.width) { x -> bitMatrix.get(x, y) }
            },
        )
    }

    fun encodeImage(content: String, size: Int): ImageBitmap {
        val matrix = encode(content = content, size = size)
        val pixels = IntArray(matrix.width * matrix.height)
        matrix.rows.forEachIndexed { y, row ->
            row.forEachIndexed { x, isDark ->
                pixels[(y * matrix.width) + x] = if (isDark) 0xFF111111.toInt() else 0xFFFFFFFF.toInt()
            }
        }
        return Bitmap.createBitmap(
            pixels,
            matrix.width,
            matrix.height,
            Bitmap.Config.ARGB_8888,
        ).asImageBitmap()
    }
}
