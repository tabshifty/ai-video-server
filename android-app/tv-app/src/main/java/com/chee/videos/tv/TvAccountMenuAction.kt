package com.chee.videos.tv

enum class TvAccountMenuAction(val label: String) {
    Repair("重新配对"),
    Logout("退出登录"),
    SwitchServer("切换服务器"),
    ;

    companion object {
        fun defaults(): List<TvAccountMenuAction> = listOf(Repair, Logout, SwitchServer)
    }
}
