package com.chee.videos.feature.connection

import com.chee.videos.core.model.ServerEndpoint
import com.chee.videos.core.repository.ConnectionServerRepository
import com.chee.videos.core.testing.MainDispatcherRule
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.runCurrent
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class ConnectionViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun manualConnectRecoversWhenEndpointProbeThrows() = runTest {
        val repository = FakeConnectionServerRepository(testError = IllegalStateException("网络异常"))
        val viewModel = ConnectionViewModel(repository)

        viewModel.updateHostInput("192.168.1.24")
        viewModel.updatePortInput("8080")
        viewModel.manualConnect()
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.connecting)
        assertTrue(state.messageIsError)
        assertEquals("连接失败：网络异常", state.message)
        assertTrue(repository.activatedEndpoints.isEmpty())
    }

    @Test
    fun useEndpointIgnoresRepeatedClicksWhileConnecting() = runTest {
        val probeGate = CompletableDeferred<Boolean>()
        val repository = FakeConnectionServerRepository(testGate = probeGate)
        val viewModel = ConnectionViewModel(repository)

        viewModel.useEndpoint("http://192.168.1.24:8080")
        runCurrent()
        viewModel.useEndpoint("http://192.168.1.25:8080")

        assertTrue(viewModel.uiState.value.connecting)
        assertEquals(listOf("http://192.168.1.24:8080"), repository.testRequests)

        probeGate.complete(true)
        advanceUntilIdle()

        assertFalse(viewModel.uiState.value.connecting)
        assertEquals(listOf("http://192.168.1.24:8080"), repository.activatedEndpoints)
    }

    @Test
    fun manualConnectShowsFailureMessageWhenEndpointProbeReturnsFalse() = runTest {
        val repository = FakeConnectionServerRepository(testResult = false)
        val viewModel = ConnectionViewModel(repository)

        viewModel.updateHostInput("192.168.1.24")
        viewModel.updatePortInput("8080")
        viewModel.manualConnect()
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.connecting)
        assertTrue(state.messageIsError)
        assertEquals("连接失败，请确认 IP/端口和服务器状态", state.message)
        assertTrue(repository.activatedEndpoints.isEmpty())
    }
}

private class FakeConnectionServerRepository(
    private val testResult: Boolean = true,
    private val testError: Throwable? = null,
    private val testGate: CompletableDeferred<Boolean>? = null,
) : ConnectionServerRepository {
    private val endpoints = MutableStateFlow<List<ServerEndpoint>>(emptyList())

    val testRequests = mutableListOf<String>()
    val activatedEndpoints = mutableListOf<String>()

    override val endpointsFlow: Flow<List<ServerEndpoint>> = endpoints

    override suspend fun scanLocalNetwork(ports: List<Int>): List<ServerEndpoint> = emptyList()

    override suspend fun testEndpoint(rawInput: String): Boolean {
        testRequests += rawInput
        testError?.let { throw it }
        return testGate?.await() ?: testResult
    }

    override suspend fun activateEndpoint(baseUrl: String, clearTokens: Boolean) {
        activatedEndpoints += baseUrl
    }

    override suspend fun removeEndpoint(baseUrl: String) = Unit
}
