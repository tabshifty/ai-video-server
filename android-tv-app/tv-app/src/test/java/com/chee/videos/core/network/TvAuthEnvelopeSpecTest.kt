package com.chee.videos.core.network

import com.chee.videos.core.model.ApiEnvelope
import com.chee.videos.core.model.TvAuthCreateEnvelope
import com.chee.videos.core.model.TvAuthSessionCreatePayload
import com.chee.videos.core.model.TvAuthSessionStatusPayload
import com.chee.videos.core.model.TvAuthStatusEnvelope
import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvAuthEnvelopeSpecTest {

    @Test
    fun `tv auth endpoints use concrete envelopes instead of generic ApiEnvelope`() {
        val apiSource = Path.of("src/main/java/com/chee/videos/core/network/ApiService.kt").readText()

        assertTrue(
            "创建配对会话端点必须返回具体 TvAuthCreateEnvelope，避免 Gson 泛型 data 退化为 LinkedTreeMap",
            apiSource.contains("): TvAuthCreateEnvelope"),
        )
        assertTrue(
            "读取配对会话端点必须返回具体 TvAuthStatusEnvelope，避免 Gson 泛型 data 退化为 LinkedTreeMap",
            apiSource.contains("): TvAuthStatusEnvelope"),
        )
        assertFalse(
            "createTvAuthSession 不能再返回 ApiEnvelope<TvAuthSessionCreatePayload>",
            apiSource.contains("): ApiEnvelope<TvAuthSessionCreatePayload>"),
        )
        assertFalse(
            "getTvAuthSession 不能再返回 ApiEnvelope<TvAuthSessionStatusPayload>",
            apiSource.contains("): ApiEnvelope<TvAuthSessionStatusPayload>"),
        )
    }

    @Test
    fun `tv auth concrete envelopes expose strongly typed data fields`() {
        val createDataField = TvAuthCreateEnvelope::class.java
            .getDeclaredField("data")
        val statusDataField = TvAuthStatusEnvelope::class.java
            .getDeclaredField("data")

        assertTrue(
            "TvAuthCreateEnvelope.data 必须是 TvAuthSessionCreatePayload",
            createDataField.type == TvAuthSessionCreatePayload::class.java,
        )
        assertTrue(
            "TvAuthStatusEnvelope.data 必须是 TvAuthSessionStatusPayload",
            statusDataField.type == TvAuthSessionStatusPayload::class.java,
        )
    }

    @Test
    fun `generic ApiEnvelope remains available for endpoints that do not consume tv auth payload data`() {
        val apiSource = Path.of("src/main/java/com/chee/videos/core/network/ApiService.kt").readText()

        assertTrue(
            "approve / deny 仍可使用泛型 ApiEnvelope，因为调用方不读取 data payload",
            apiSource.contains("): ApiEnvelope<Map<String, Boolean>>"),
        )
        assertTrue(ApiEnvelope(code = 0, data = mapOf("ok" to true)).data?.get("ok") == true)
    }

    @Test
    fun `release shrinker keeps tv auth gson models used through retrofit signatures`() {
        val proguardRules = Path.of("proguard-rules.pro").readText()

        assertTrue(
            "Release R8 必须保留 TV 授权模型；否则 suspend Retrofit 返回类型只在 Signature 中出现时会被 shrink，设备上退化为错误类型",
            proguardRules.contains("-keep class com.chee.videos.core.model.TvAuth* { *; }"),
        )
    }
}
