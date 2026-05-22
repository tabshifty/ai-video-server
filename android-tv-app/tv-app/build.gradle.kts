plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
    id("org.jetbrains.kotlin.kapt")
    id("com.google.dagger.hilt.android")
}

val tvMainSourceExcludes = listOf(
    "com/chee/videos/MainActivity.kt",
    "com/chee/videos/VideoHomeApp.kt",
    "com/chee/videos/feature/auth/**",
    "com/chee/videos/feature/home/**",
    "com/chee/videos/feature/mine/**",
    "com/chee/videos/feature/player/**",
    "com/chee/videos/feature/shorts/**",
    "com/chee/videos/feature/shortdiscover/**",
    "com/chee/videos/feature/shortsearch/**",
    "com/chee/videos/feature/imagecollections/**",
    "com/chee/videos/core/ui/ShortVideo*",
)

val tvTestSourceExcludes = listOf(
    "com/chee/videos/feature/shorts/**",
    "com/chee/videos/feature/shortdiscover/**",
    "com/chee/videos/feature/shortsearch/**",
    "com/chee/videos/feature/imagecollections/**",
    "com/chee/videos/feature/home/**",
    "com/chee/videos/feature/player/**",
    "com/chee/videos/core/ui/ShortVideo*",
)

android {
    namespace = "com.chee.videos.tv"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.chee.videos.tv"
        minSdk = 26
        targetSdk = 35
        versionCode = 61
        versionName = "0.1.60"
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
        vectorDrawables {
            useSupportLibrary = true
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro",
            )
        }
    }

    splits {
        abi {
            isEnable = true
            reset()
            include("armeabi-v7a", "arm64-v8a")
            exclude("x86", "x86_64")
            isUniversalApk = false
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    kotlinOptions {
        jvmTarget = "17"
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }
    composeOptions {
        kotlinCompilerExtensionVersion = "1.5.15"
    }
    packaging {
        resources {
            excludes += "/META-INF/{AL2.0,LGPL2.1}"
        }
    }

}

kapt {
    correctErrorTypes = true
}

kotlin {
    sourceSets {
        getByName("main") {
            tvMainSourceExcludes.forEach { kotlin.exclude(it) }
        }
        getByName("test") {
            tvTestSourceExcludes.forEach { kotlin.exclude(it) }
        }
    }
}

tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile>().configureEach {
    val excludes = if (name.contains("UnitTest", ignoreCase = true)) {
        tvTestSourceExcludes
    } else {
        tvMainSourceExcludes
    }
    exclude(excludes)
}

tasks.withType<org.jetbrains.kotlin.gradle.internal.KaptGenerateStubsTask>().configureEach {
    val excludes = if (name.contains("UnitTest", ignoreCase = true)) {
        tvTestSourceExcludes
    } else {
        tvMainSourceExcludes
    }
    exclude(excludes)
}

dependencies {
    val composeBom = platform("androidx.compose:compose-bom:2024.10.01")
    implementation(composeBom)
    androidTestImplementation(composeBom)

    implementation("androidx.core:core-ktx:1.13.1")
    implementation("androidx.lifecycle:lifecycle-runtime-ktx:2.8.4")
    implementation("androidx.lifecycle:lifecycle-runtime-compose:2.8.4")
    implementation("androidx.activity:activity-compose:1.9.1")
    implementation("androidx.navigation:navigation-compose:2.7.7")
    implementation("androidx.lifecycle:lifecycle-viewmodel-compose:2.8.4")

    implementation("androidx.compose.ui:ui")
    implementation("androidx.compose.ui:ui-graphics")
    implementation("androidx.compose.ui:ui-tooling-preview")
    implementation("androidx.compose.material:material")
    implementation("androidx.compose.material3:material3")
    implementation("androidx.compose.material:material-icons-extended")
    implementation("com.google.android.material:material:1.12.0")
    debugImplementation("androidx.compose.ui:ui-tooling")
    debugImplementation("androidx.compose.ui:ui-test-manifest")

    implementation("androidx.datastore:datastore-preferences:1.1.1")

    implementation("com.google.dagger:hilt-android:2.51.1")
    kapt("com.google.dagger:hilt-compiler:2.51.1")
    implementation("androidx.hilt:hilt-navigation-compose:1.2.0")

    implementation("com.squareup.retrofit2:retrofit:2.11.0")
    implementation("com.squareup.retrofit2:converter-gson:2.11.0")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("com.squareup.okhttp3:logging-interceptor:4.12.0")

    implementation("io.coil-kt:coil-compose:2.7.0")
    implementation("com.google.zxing:core:3.5.3")

    implementation("androidx.media3:media3-exoplayer:1.4.1")
    implementation("androidx.media3:media3-exoplayer-hls:1.4.1")
    implementation("androidx.media3:media3-ui:1.4.1")
    implementation("org.videolan.android:libvlc-all:3.6.0")

    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.8.1")
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.7.1")

    testImplementation("junit:junit:4.13.2")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.8.1")
    androidTestImplementation("androidx.test.ext:junit:1.2.1")
    androidTestImplementation("androidx.test.espresso:espresso-core:3.6.1")
    androidTestImplementation("androidx.compose.ui:ui-test-junit4")
}
