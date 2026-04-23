package com.chee.videos.feature.imagecollections

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.gestures.detectTransformGestures
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.lazy.staggeredgrid.LazyVerticalStaggeredGrid
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridCells
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridItemSpan
import androidx.compose.foundation.lazy.staggeredgrid.itemsIndexed as staggeredItemsIndexed
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Collections
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.listSaver
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.chee.videos.core.model.ImageCollectionDetailDto
import com.chee.videos.core.model.ImageCollectionListItemDto
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.util.UrlBuilder
import kotlin.math.absoluteValue
import kotlin.math.max
import kotlin.math.min
import kotlinx.coroutines.launch

@Composable
fun ImageCollectionsScreen(
    baseUrl: String,
    onBack: (() -> Unit)?,
    onOpenCollection: (String) -> Unit,
    viewModel: ImageCollectionsViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.load()
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient)
            .statusBarsPadding(),
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            ImageCollectionsTopBar(
                title = "图片合集",
                subtitle = if (normalizeImageCollectionsQuery(uiState.query) == null) {
                    if (uiState.totalCount > 0) "共 ${uiState.totalCount} 个合集" else "打开合集后可沉浸式浏览原图"
                } else {
                    "按标题搜索图片合集"
                },
                onBack = onBack,
            )
            ImageCollectionsSearchBar(
                query = uiState.query,
                onQueryChanged = viewModel::onQueryChanged,
            )

            when {
                uiState.loading && uiState.items.isEmpty() -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = AppChrome.AccentStrong)
                    }
                }

                !uiState.errorMessage.isNullOrBlank() && uiState.items.isEmpty() -> {
                    ErrorState(
                        message = uiState.errorMessage.orEmpty(),
                        action = "重新加载",
                        onAction = viewModel::retry,
                    )
                }

                uiState.items.isEmpty() -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        Text(imageCollectionsEmptyMessage(uiState.query), color = AppChrome.TextSecondary)
                    }
                }

                else -> {
                    LazyVerticalStaggeredGrid(
                        columns = StaggeredGridCells.Fixed(2),
                        modifier = Modifier.fillMaxSize(),
                        contentPadding = PaddingValues(start = 14.dp, end = 14.dp, top = 6.dp, bottom = 24.dp),
                        horizontalArrangement = Arrangement.spacedBy(12.dp),
                        verticalItemSpacing = 12.dp,
                    ) {
                        staggeredItemsIndexed(uiState.items, key = { _, item -> item.id }) { index, item ->
                            if (index >= uiState.items.lastIndex - 5) {
                                LaunchedEffect(index, uiState.items.size, uiState.loadingMore) {
                                    viewModel.loadMoreIfNeeded(index)
                                }
                            }
                            ImageCollectionCard(
                                baseUrl = baseUrl,
                                item = item,
                                onClick = { onOpenCollection(item.id) },
                            )
                        }
                        if (uiState.loadingMore) {
                            item(span = StaggeredGridItemSpan.FullLine) {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(vertical = 8.dp),
                                    horizontalArrangement = Arrangement.Center,
                                    verticalAlignment = Alignment.CenterVertically,
                                ) {
                                    CircularProgressIndicator(
                                        modifier = Modifier.size(18.dp),
                                        color = AppChrome.AccentStrong,
                                        strokeWidth = 2.dp,
                                    )
                                    Text(
                                        text = "正在加载更多图集",
                                        color = AppChrome.TextMuted,
                                        style = MaterialTheme.typography.bodySmall,
                                        modifier = Modifier.padding(start = 8.dp),
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun ImageCollectionViewerScreen(
    baseUrl: String,
    onBack: () -> Unit,
    viewModel: ImageCollectionViewerViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val detail = uiState.detail

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        when {
            uiState.loading -> {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = AppChrome.AccentStrong)
                }
            }

            !uiState.errorMessage.isNullOrBlank() || detail == null -> {
                ErrorState(
                    message = uiState.errorMessage ?: "图集详情加载失败",
                    action = "重新加载",
                    onAction = viewModel::load,
                )
            }

            else -> {
                ImageCollectionViewerContent(
                    baseUrl = baseUrl,
                    detail = detail,
                    onBack = onBack,
                )
            }
        }
    }
}

@Composable
private fun ImageCollectionsTopBar(
    title: String,
    subtitle: String,
    onBack: (() -> Unit)?,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 8.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        if (onBack != null) {
            IconButton(onClick = onBack) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回", tint = AppChrome.TextPrimary)
            }
        } else {
            Box(modifier = Modifier.size(16.dp))
        }
        Column(
            modifier = Modifier.weight(1f),
            verticalArrangement = Arrangement.spacedBy(2.dp),
        ) {
            Text(
                text = title,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = subtitle,
                color = AppChrome.TextMuted,
                style = MaterialTheme.typography.bodySmall,
            )
        }
    }
}

@Composable
private fun ImageCollectionsSearchBar(
    query: String,
    onQueryChanged: (String) -> Unit,
) {
    OutlinedTextField(
        value = query,
        onValueChange = onQueryChanged,
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        singleLine = true,
        placeholder = {
            Text("搜索图集标题", color = AppChrome.TextMuted)
        },
        leadingIcon = {
            Icon(Icons.Filled.Search, contentDescription = null, tint = AppChrome.TextMuted)
        },
        trailingIcon = {
            if (query.isNotBlank()) {
                IconButton(onClick = { onQueryChanged("") }) {
                    Icon(Icons.Filled.Close, contentDescription = "清空搜索", tint = AppChrome.TextMuted)
                }
            }
        },
        shape = RoundedCornerShape(18.dp),
        colors = OutlinedTextFieldDefaults.colors(
            focusedTextColor = AppChrome.TextPrimary,
            unfocusedTextColor = AppChrome.TextPrimary,
            focusedBorderColor = AppChrome.AccentStrong,
            unfocusedBorderColor = AppChrome.Divider,
            cursorColor = AppChrome.AccentStrong,
            focusedContainerColor = AppChrome.Surface.copy(alpha = 0.92f),
            unfocusedContainerColor = AppChrome.Surface.copy(alpha = 0.82f),
        ),
    )
}

@Composable
private fun ImageCollectionCard(
    baseUrl: String,
    item: ImageCollectionListItemDto,
    onClick: () -> Unit,
) {
    val ratio = remember(item.id) { imageCollectionCardRatio(item.id) }
    val coverUrl = remember(baseUrl, item.coverUrl) { resolveImageUrl(baseUrl, item.coverUrl) }

    Surface(
        color = AppChrome.SurfaceElevated,
        shape = RoundedCornerShape(18.dp),
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(18.dp))
            .clickable(onClick = onClick),
    ) {
        Box {
            if (coverUrl.isNullOrBlank()) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .aspectRatio(ratio)
                        .background(AppChrome.SurfaceStrong),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        imageVector = Icons.Filled.Collections,
                        contentDescription = null,
                        tint = AppChrome.TextMuted,
                        modifier = Modifier.size(30.dp),
                    )
                }
            } else {
                AsyncImage(
                    model = coverUrl,
                    contentDescription = item.name,
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .fillMaxWidth()
                        .aspectRatio(ratio),
                )
            }
            Box(
                modifier = Modifier
                    .matchParentSize()
                    .background(
                        Brush.verticalGradient(
                            colors = listOf(Color.Transparent, Color.Transparent, Color(0xD9000000)),
                        ),
                    ),
            )
            Column(
                modifier = Modifier
                    .align(Alignment.BottomStart)
                    .padding(horizontal = 12.dp, vertical = 12.dp),
                verticalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                Text(
                    text = item.name,
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = "${item.imageCount.coerceAtLeast(0)} 张图片",
                    color = AppChrome.TextSecondary,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun ImageCollectionViewerContent(
    baseUrl: String,
    detail: ImageCollectionDetailDto,
    onBack: () -> Unit,
) {
    val images = detail.images
    if (images.isEmpty()) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            Text("合集里还没有可浏览的图片", color = AppChrome.TextSecondary)
        }
        return
    }

    val pagerState = rememberPagerState(pageCount = { images.size })
    val thumbState = rememberLazyListState()
    val scope = androidx.compose.runtime.rememberCoroutineScope()
    var chromeVisible by rememberSaveable(detail.id) { mutableStateOf(true) }
    var imageScale by rememberSaveable(detail.id) { mutableStateOf(1f) }
    var imageOffset by rememberSaveable(detail.id, stateSaver = ImageViewerOffsetSaver) { mutableStateOf(Offset.Zero) }

    LaunchedEffect(pagerState.currentPage) {
        thumbState.animateScrollToItem(pagerState.currentPage)
        imageScale = 1f
        imageOffset = Offset.Zero
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        BoxWithConstraints(modifier = Modifier.fillMaxSize()) {
            val density = LocalDensity.current
            val containerWidthPx = with(density) { maxWidth.toPx() }
            val containerHeightPx = with(density) { maxHeight.toPx() }
            val currentImage = images[pagerState.currentPage]
            val maxScale = remember(currentImage.id, containerWidthPx, containerHeightPx) {
                imageViewerMaxScale(
                    imageWidthPx = currentImage.width.toFloat(),
                    imageHeightPx = currentImage.height.toFloat(),
                    containerWidthPx = containerWidthPx,
                    containerHeightPx = containerHeightPx,
                )
            }
            val displayScale = imageScale.coerceIn(ImageViewerMinScale, maxScale)
            val displayOffset = imageViewerClampOffset(
                offset = imageOffset,
                scale = displayScale,
                imageWidthPx = currentImage.width.toFloat(),
                imageHeightPx = currentImage.height.toFloat(),
                containerWidthPx = containerWidthPx,
                containerHeightPx = containerHeightPx,
            )

            HorizontalPager(
                state = pagerState,
                userScrollEnabled = displayScale <= 1.001f,
                modifier = Modifier.fillMaxSize(),
            ) { page ->
                val image = images[page]
                val viewUrl = remember(baseUrl, image.viewUrl) { resolveImageUrl(baseUrl, image.viewUrl) }
                val isCurrentPage = page == pagerState.currentPage
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .pointerInput(chromeVisible) {
                            detectTapGestures(
                                onDoubleTap = {
                                    chromeVisible = toggleImageCollectionViewerChrome(chromeVisible)
                                },
                            )
                        }
                        .pointerInput(
                            image.id,
                            isCurrentPage,
                            maxScale,
                            containerWidthPx,
                            containerHeightPx,
                            imageScale,
                            imageOffset,
                        ) {
                            if (!isCurrentPage) {
                                return@pointerInput
                            }
                            detectTransformGestures { _, pan, zoom, _ ->
                                val baseScale = imageScale.coerceIn(ImageViewerMinScale, maxScale)
                                val baseOffset = imageViewerClampOffset(
                                    offset = imageOffset,
                                    scale = baseScale,
                                    imageWidthPx = image.width.toFloat(),
                                    imageHeightPx = image.height.toFloat(),
                                    containerWidthPx = containerWidthPx,
                                    containerHeightPx = containerHeightPx,
                                )
                                val nextScale = (baseScale * zoom).coerceIn(ImageViewerMinScale, maxScale)
                                val rawOffset = if (nextScale <= 1f) {
                                    Offset.Zero
                                } else {
                                    baseOffset + pan
                                }
                                imageScale = nextScale
                                imageOffset = imageViewerClampOffset(
                                    offset = rawOffset,
                                    scale = nextScale,
                                    imageWidthPx = image.width.toFloat(),
                                    imageHeightPx = image.height.toFloat(),
                                    containerWidthPx = containerWidthPx,
                                    containerHeightPx = containerHeightPx,
                                )
                            }
                        },
                    contentAlignment = Alignment.Center,
                ) {
                    AsyncImage(
                        model = viewUrl,
                        contentDescription = image.title,
                        contentScale = ContentScale.Fit,
                        modifier = Modifier
                            .fillMaxSize()
                            .graphicsLayer {
                                if (isCurrentPage) {
                                    scaleX = displayScale
                                    scaleY = displayScale
                                    translationX = displayOffset.x
                                    translationY = displayOffset.y
                                } else {
                                    scaleX = 1f
                                    scaleY = 1f
                                    translationX = 0f
                                    translationY = 0f
                                }
                            },
                    )
                }
            }
        }

        AnimatedVisibility(
            visible = chromeVisible,
            enter = fadeIn(),
            exit = fadeOut(),
            modifier = Modifier.align(Alignment.TopCenter),
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(
                        Brush.verticalGradient(
                            colors = listOf(Color(0xD9000000), Color.Transparent),
                        ),
                    )
                    .statusBarsPadding()
                    .padding(horizontal = 8.dp, vertical = 8.dp),
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    Surface(
                        color = Color(0x4DFFFFFF),
                        shape = CircleShape,
                    ) {
                        IconButton(onClick = onBack) {
                            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回", tint = Color.White)
                        }
                    }
                    Column(
                        modifier = Modifier.weight(1f),
                        verticalArrangement = Arrangement.spacedBy(2.dp),
                    ) {
                        Text(
                            text = detail.name,
                            color = Color.White,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis,
                        )
                        Text(
                            text = "${pagerState.currentPage + 1}/${images.size}",
                            color = Color(0xFFD6D9E1),
                            style = MaterialTheme.typography.bodySmall,
                        )
                    }
                }
            }
        }

        AnimatedVisibility(
            visible = chromeVisible,
            enter = fadeIn(),
            exit = fadeOut(),
            modifier = Modifier.align(Alignment.BottomCenter),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(
                        Brush.verticalGradient(
                            colors = listOf(Color.Transparent, Color(0xE0000000)),
                        ),
                    )
                    .navigationBarsPadding()
                    .padding(horizontal = 12.dp, vertical = 14.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                val currentImage = images[pagerState.currentPage]
                if (currentImage.title.isNotBlank()) {
                    Text(
                        text = currentImage.title,
                        color = Color.White,
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.SemiBold,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
                LazyRow(
                    state = thumbState,
                    horizontalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    itemsIndexed(images, key = { _, image -> image.id }) { index, image ->
                        val thumbUrl = remember(baseUrl, image.thumbnailUrl) {
                            resolveImageUrl(baseUrl, image.thumbnailUrl)
                        }
                        val selected = index == pagerState.currentPage
                        Surface(
                            color = if (selected) Color(0x4DFFFFFF) else Color(0x1AFFFFFF),
                            shape = RoundedCornerShape(14.dp),
                            border = androidx.compose.foundation.BorderStroke(
                                width = if (selected) 1.5.dp else 1.dp,
                                color = if (selected) AppChrome.AccentStrong else Color.Transparent,
                            ),
                            modifier = Modifier
                                .size(width = 74.dp, height = 92.dp)
                                .clip(RoundedCornerShape(14.dp))
                                .clickable {
                                    if (index == pagerState.currentPage) {
                                        imageScale = 1f
                                        imageOffset = Offset.Zero
                                    } else {
                                        scope.launch {
                                            pagerState.animateScrollToPage(index)
                                        }
                                    }
                                },
                        ) {
                            if (!thumbUrl.isNullOrBlank()) {
                                AsyncImage(
                                    model = thumbUrl,
                                    contentDescription = image.title,
                                    contentScale = ContentScale.Crop,
                                    modifier = Modifier.fillMaxSize(),
                                )
                            } else {
                                Box(
                                    modifier = Modifier
                                        .fillMaxSize()
                                        .background(Color(0x33000000)),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    Icon(
                                        imageVector = Icons.Filled.Collections,
                                        contentDescription = null,
                                        tint = Color.White.copy(alpha = 0.6f),
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ErrorState(
    message: String,
    action: String,
    onAction: () -> Unit,
) {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Surface(
            color = AppChrome.SurfaceElevated,
            shape = AppChrome.CardShape,
        ) {
            Column(
                modifier = Modifier.padding(horizontal = 20.dp, vertical = 18.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text(message, color = AppChrome.Error)
                Surface(
                    color = AppChrome.AccentSoft,
                    shape = AppChrome.PillShape,
                    modifier = Modifier.clickable(onClick = onAction),
                ) {
                    Text(
                        text = action,
                        color = AppChrome.TextPrimary,
                        style = MaterialTheme.typography.labelLarge,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 10.dp),
                    )
                }
            }
        }
    }
}

private fun imageCollectionCardRatio(collectionId: String): Float {
    return when (collectionId.hashCode().absoluteValue % 3) {
        0 -> 0.78f
        1 -> 0.92f
        else -> 1.06f
    }
}

private fun resolveImageUrl(baseUrl: String, rawPath: String?): String? {
    val path = rawPath?.trim().orEmpty()
    if (path.isBlank()) {
        return null
    }
    if (path.startsWith("http://") || path.startsWith("https://")) {
        return path
    }
    return UrlBuilder.normalizeBaseUrl(baseUrl) + path
}

internal const val ImageViewerMinScale = 0.6f
private const val ImageViewerMaxScaleCap = 4f
internal val ImageViewerOffsetSaver = listSaver<Offset, Float>(
    save = { offset -> saveImageViewerOffset(offset) },
    restore = { values -> restoreImageViewerOffset(values) },
)

internal fun saveImageViewerOffset(offset: Offset): List<Float> = listOf(offset.x, offset.y)

internal fun restoreImageViewerOffset(values: List<Float>): Offset {
    if (values.size < 2) {
        return Offset.Zero
    }
    return Offset(values[0], values[1])
}

internal fun imageViewerMaxScale(
    imageWidthPx: Float,
    imageHeightPx: Float,
    containerWidthPx: Float,
    containerHeightPx: Float,
): Float {
    if (imageWidthPx <= 0f || imageHeightPx <= 0f || containerWidthPx <= 0f || containerHeightPx <= 0f) {
        return 1f
    }
    if (imageWidthPx <= containerWidthPx && imageHeightPx <= containerHeightPx) {
        return 1f
    }
    val ratio = max(imageWidthPx / containerWidthPx, imageHeightPx / containerHeightPx)
    return ratio.coerceIn(1f, ImageViewerMaxScaleCap)
}

internal fun imageViewerClampOffset(
    offset: Offset,
    scale: Float,
    imageWidthPx: Float,
    imageHeightPx: Float,
    containerWidthPx: Float,
    containerHeightPx: Float,
): Offset {
    if (scale <= 1f || imageWidthPx <= 0f || imageHeightPx <= 0f || containerWidthPx <= 0f || containerHeightPx <= 0f) {
        return Offset.Zero
    }
    val baseSize = imageViewerFitSize(
        imageWidthPx = imageWidthPx,
        imageHeightPx = imageHeightPx,
        containerWidthPx = containerWidthPx,
        containerHeightPx = containerHeightPx,
    )
    val scaledWidth = baseSize.first * scale
    val scaledHeight = baseSize.second * scale
    val maxOffsetX = max(0f, (scaledWidth - containerWidthPx) / 2f)
    val maxOffsetY = max(0f, (scaledHeight - containerHeightPx) / 2f)
    return Offset(
        x = offset.x.coerceIn(-maxOffsetX, maxOffsetX),
        y = offset.y.coerceIn(-maxOffsetY, maxOffsetY),
    )
}

private fun imageViewerFitSize(
    imageWidthPx: Float,
    imageHeightPx: Float,
    containerWidthPx: Float,
    containerHeightPx: Float,
): Pair<Float, Float> {
    val imageRatio = imageWidthPx / imageHeightPx
    val containerRatio = containerWidthPx / containerHeightPx
    return if (imageRatio >= containerRatio) {
        val width = containerWidthPx
        val height = width / imageRatio
        width to min(height, containerHeightPx)
    } else {
        val height = containerHeightPx
        val width = height * imageRatio
        min(width, containerWidthPx) to height
    }
}
