package com.chee.videos.feature.tvauth

import com.journeyapps.barcodescanner.CaptureActivity

/**
 * 扫码登录 TV 用的自定义 CaptureActivity。
 *
 * zxing-android-embedded 默认的 CaptureActivity 在其 Manifest 里声明了 sensorLandscape，
 * 若不注册自定义子类覆盖，扫码页会被锁成横屏；setOrientationLocked(false) 只抑制运行时
 * setRequestedOrientation，清不掉 Manifest 级方向锁。这里用空子类 + 在 app Manifest 里
 * 显式锁定竖屏，让方向完全由 app 控制。
 */
class PortraitCaptureActivity : CaptureActivity()
