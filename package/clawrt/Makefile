include $(TOPDIR)/rules.mk

PKG_NAME:=clawrt
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

PKG_SOURCE_PROTO:=git
PKG_SOURCE_URL:=https://github.com/Havr88/clawrt.git
PKG_SOURCE_VERSION:=main

PKG_MAINTAINER:=ClawRT Contributors <dev@clawrt.local>
PKG_LICENSE:=MIT
PKG_LICENSE_FILES:=LICENSE

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

GO_PKG:=clawrt/cmd/clawrt

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/clawrt
  SECTION:=net
  CATEGORY:=Network
  TITLE:=ClawRT AI Router Agent for OpenWrt
  URL:=https://github.com/Havr88/clawrt
  DEPENDS:=$(GO_ARCH_DEPENDS) +ca-certificates
endef

define Package/clawrt/description
  ClawRT is a lightweight, secure agentic AI daemon tailored for OpenWrt routers.
endef

define Package/clawrt/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(GO_PKG_BUILD_BIN_DIR)/clawrt $(1)/usr/bin/clawrt
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/clawrt.config $(1)/etc/config/clawrt
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/clawrt.init $(1)/etc/init.d/clawrt
endef

$(eval $(call BuildPackage,clawrt))
