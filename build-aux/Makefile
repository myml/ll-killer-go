.PHONY: all create apt-update-pkg apt-install extra-apt-install commit export config clean clean-ll clean-fs clean-apt

-include config.mk

PKG ?= app
APPID ?= $(PKG).linyaps
BASE ?= 
ifeq ($(PKG),)
  $(error "请指定PKG变量，如make PKG=APPID")
endif
ifeq ($(PKG),)
  $(error "请指定BASE变量，如make BASE=org.deepin.base/23.1.0")
endif

FUSE := $(PWD)/fuse-overlayfs
CACHE_FILE := linglong/apt/cache/apt/pkgcache.bin
KL ?= ll-killer
export KILLER_EXEC ?= $(KL)
export KILLER_BUILD_ARGS ?= --fuse-overlayfs fuse-overlayfs
KILLER_CREATE_ARGS ?= 


all: export
create: $(PKG)/linglong.yaml
apt-update-pkg: $(PKG)/apt-update.log
apt-install-extra: $(PKG)/apt-install-extra.log
config: config.mk

config.mk:
	@echo "[生成config.mk]"
	@echo "PKG ?= $(PKG)" > config.mk
	@echo "APPID ?= $(APPID)" >> config.mk
	@echo "KILLER_EXEC ?= $(KILLER_EXEC)" >> config.mk
	@echo "BASE ?= $(BASE)" >> config.mk
	@echo "KILLER_BUILD_ARGS ?= $(KILLER_BUILD_ARGS)" >> config.mk
	@echo "KILLER_CREATE_ARGS ?= $(KILLER_CREATE_ARGS)" >> config.mk

apt.conf.d/ll-killer.conf: config.mk
	@echo "[创建辅助脚本]"
	$(KILLER_EXEC) build-aux

$(CACHE_FILE): sources.list apt.conf.d/ll-killer.conf
	@echo "[更新APT缓存]"
	$(KILLER_EXEC) apt -- apt update -y

pkg.info: $(CACHE_FILE)
	@echo "[获取包元数据]"
	$(KILLER_EXEC) apt -- apt show $(PKG) > pkg.info

linglong.yaml deps.list: pkg.info
	@echo "[创建玲珑项目]"
	$(KILLER_EXEC) create --from pkg.info --id "$(APPID)" $(KILLER_CREATE_ARGS) 
	touch deps.list

apt-install.log: $(CACHE_FILE) deps.list
	@echo "[将包安装至构建环境]"
	$(KILLER_EXEC) build $(KILLER_BUILD_ARGS) --ptrace -- apt install -y $(PKG) $(shell cat deps.list) | tee apt-install.log

ldd-check.log: apt-install.log
	@echo "[检查缺失库]"
	$(KILLER_EXEC) build $(KILLER_BUILD_ARGS) -- build-aux/ldd-check.sh > ldd-check.log

ldd-found.log ldd-notfound.log: ldd-check.log
	@echo "[搜索缺失库所在包]"
	$(KILLER_EXEC) apt -- build-aux/ldd-search.sh ldd-check.log ldd-found.log ldd-notfound.log;

apt-install-extra.log: apt-install.log ldd-found.log
	@echo "[安装找到的缺失库]"
	$(KILLER_EXEC) build -- apt install -y $(shell cat ldd-found.log) | tee apt-install-extra.log;

commit: apt-install-extra.log
	@echo "[提交文件到容器中]"
	$(KILLER_EXEC) commit -- --skip-output-check --skip-strip-symbols

export: commit
	@echo "[导出layer文件]"
	$(KILLER_EXEC) export -- --layer

clean-apt:
	rm -rf linglong/apt
clean-fs:
	rm -rf linglong/filesystem
clean-ll:
	rm -rf linglong/output
clean: clean-ll clean-fs
	rm -rf *.log *.layer