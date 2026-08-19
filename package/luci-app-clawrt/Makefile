include $(TOPDIR)/rules.mk

LUCI_TITLE:=LuCI support for ClawRT AI Router Agent
LUCI_DEPENDS:=+clawrt
LUCI_PKGARCH:=all

PKG_NAME:=luci-app-clawrt
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

PKG_MAINTAINER:=ClawRT Contributors <dev@clawrt.local>
PKG_LICENSE:=MIT

include $(TOPDIR)/feeds/luci/luci.mk

# call BuildPackage - OpenWrt buildroot signature
$(eval $(call luci.mk))
