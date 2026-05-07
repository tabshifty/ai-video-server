package com.chee.videos.core.di

import com.chee.videos.feature.tv.NetworkTvRepository
import com.chee.videos.feature.tv.TvRepository
import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
abstract class TvRepositoryModule {
    @Binds
    @Singleton
    abstract fun bindTvRepository(impl: NetworkTvRepository): TvRepository
}
