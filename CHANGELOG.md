# Changelog

## [0.16.3](https://github.com/rudderlabs/rudder-go-kit/compare/v0.16.2...v0.16.3) (2023-10-30)


### Bug Fixes

* soft failure for invalid instruments ([#181](https://github.com/rudderlabs/rudder-go-kit/issues/181)) ([0f91858](https://github.com/rudderlabs/rudder-go-kit/commit/0f918581ab6c046cf6eae3362076c259c4f162c3))

## [0.16.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.16.1...v0.16.2) (2023-10-27)


### Bug Fixes

* memory gcra limiter inconsistency when rate &lt; cost < burst ([#177](https://github.com/rudderlabs/rudder-go-kit/issues/177)) ([cda0d64](https://github.com/rudderlabs/rudder-go-kit/commit/cda0d6462a6ad7a33f54ffc018bb84b66b3d79e6))


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.46.3 to 1.46.4 ([#178](https://github.com/rudderlabs/rudder-go-kit/issues/178)) ([73dde37](https://github.com/rudderlabs/rudder-go-kit/commit/73dde37fa769c8e4931761e8a8c58a712f302d0a))

## [0.16.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.16.0...v0.16.1) (2023-10-26)


### Miscellaneous

* add a docker resource for minio ([#169](https://github.com/rudderlabs/rudder-go-kit/issues/169)) ([5ce4b79](https://github.com/rudderlabs/rudder-go-kit/commit/5ce4b79a2b4c9401faed07e28f914552fc506fad))
* updating pr template ([#175](https://github.com/rudderlabs/rudder-go-kit/issues/175)) ([af9bd26](https://github.com/rudderlabs/rudder-go-kit/commit/af9bd261c82da73247efbd058b2406189c3e1f20))

## [0.16.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.13...v0.16.0) (2023-10-26)


### Features

* upgrade opentelemetry package ([#150](https://github.com/rudderlabs/rudder-go-kit/issues/150)) ([e2b933c](https://github.com/rudderlabs/rudder-go-kit/commit/e2b933c96b5a1730e467b1e6918b9da3c937ade9))


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.45.24 to 1.45.27 ([#159](https://github.com/rudderlabs/rudder-go-kit/issues/159)) ([d645e9e](https://github.com/rudderlabs/rudder-go-kit/commit/d645e9ea63c44917d131a7b7ceca9a1faae6dea5))
* **deps:** bump github.com/aws/aws-sdk-go from 1.45.27 to 1.46.3 ([#172](https://github.com/rudderlabs/rudder-go-kit/issues/172)) ([a6e1b13](https://github.com/rudderlabs/rudder-go-kit/commit/a6e1b1370ba684e2ca10a200e2d981dd72ec539a))
* **deps:** bump github.com/fsnotify/fsnotify from 1.6.0 to 1.7.0 ([#167](https://github.com/rudderlabs/rudder-go-kit/issues/167)) ([6cbe667](https://github.com/rudderlabs/rudder-go-kit/commit/6cbe6676a820db42c8bce3c4b7d7723830ab50b5))
* **deps:** bump github.com/prometheus/client_model from 0.4.1-0.20230718164431-9a2bf3000d16 to 0.5.0 ([#163](https://github.com/rudderlabs/rudder-go-kit/issues/163)) ([3e0196f](https://github.com/rudderlabs/rudder-go-kit/commit/3e0196f1dbda90d9289ccefa560d9133de2cd6c8))
* **deps:** bump github.com/prometheus/common from 0.44.0 to 0.45.0 ([#161](https://github.com/rudderlabs/rudder-go-kit/issues/161)) ([85c4178](https://github.com/rudderlabs/rudder-go-kit/commit/85c417859d3f0fcdfb0ce0cf91aa374d73dd3c28))
* **deps:** bump google.golang.org/api from 0.147.0 to 0.148.0 ([#166](https://github.com/rudderlabs/rudder-go-kit/issues/166)) ([d93b31c](https://github.com/rudderlabs/rudder-go-kit/commit/d93b31cd957a15fe56fc86af6586417f309c7dae))
* removing unused parameter ([#173](https://github.com/rudderlabs/rudder-go-kit/issues/173)) ([be0f632](https://github.com/rudderlabs/rudder-go-kit/commit/be0f6328371715c12b5f322cbf7031123793579b))

## [0.15.13](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.12...v0.15.13) (2023-10-18)


### Miscellaneous

* change default postgres container shm size to 128MB ([6ba6edb](https://github.com/rudderlabs/rudder-go-kit/commit/6ba6edb8712eb656431d4f628ec7663877e164c3))
* code formatting ([ad42e7a](https://github.com/rudderlabs/rudder-go-kit/commit/ad42e7a5d59f0d4c5de4d6b2b79e6f4b48f08bc6))
* **deps:** bump github.com/aws/aws-sdk-go from 1.45.3 to 1.45.24 ([#144](https://github.com/rudderlabs/rudder-go-kit/issues/144)) ([f2938d3](https://github.com/rudderlabs/rudder-go-kit/commit/f2938d394e8d87ae445beb2f1ca9c666a9ec46b6))
* **deps:** bump github.com/prometheus/client_golang from 1.15.1 to 1.17.0 ([#158](https://github.com/rudderlabs/rudder-go-kit/issues/158)) ([97f7469](https://github.com/rudderlabs/rudder-go-kit/commit/97f7469163b5a189691aca1f22cc1bcaed9b415e))
* **deps:** bump github.com/spf13/viper from 1.16.0 to 1.17.0 ([#156](https://github.com/rudderlabs/rudder-go-kit/issues/156)) ([35f3f81](https://github.com/rudderlabs/rudder-go-kit/commit/35f3f815f8d6cdbe16f369673118e145997c337d))
* **deps:** bump google.golang.org/api from 0.146.0 to 0.147.0 ([#151](https://github.com/rudderlabs/rudder-go-kit/issues/151)) ([91bf8f9](https://github.com/rudderlabs/rudder-go-kit/commit/91bf8f9664c769d7c103ac06e1869321c1044e2c))

## [0.15.11](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.10...v0.15.11) (2023-09-15)


### Bug Fixes

* another config deadlock by trying to acquire a read lock twice ([#120](https://github.com/rudderlabs/rudder-go-kit/issues/120)) ([95ddd6f](https://github.com/rudderlabs/rudder-go-kit/commit/95ddd6f0bab129672e863bab8bdc4e506f25f9e8))

## [0.15.10](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.9...v0.15.10) (2023-09-15)


### Bug Fixes

* config deadlock by trying to acquire a read lock twice ([#118](https://github.com/rudderlabs/rudder-go-kit/issues/118)) ([458381e](https://github.com/rudderlabs/rudder-go-kit/commit/458381e5d1121e776c0071f0e94f2c4d92ca72a7))

## [0.15.9](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.8...v0.15.9) (2023-09-13)


### Bug Fixes

* register reloadable config variables ([#108](https://github.com/rudderlabs/rudder-go-kit/issues/108)) ([2466840](https://github.com/rudderlabs/rudder-go-kit/commit/24668400448347a63bd01a34d30440186af5141c))

## [0.15.8](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.7...v0.15.8) (2023-09-11)


### Miscellaneous

* configurable shared memory size for postgres container ([#113](https://github.com/rudderlabs/rudder-go-kit/issues/113)) ([ff2dc63](https://github.com/rudderlabs/rudder-go-kit/commit/ff2dc63b0da5151272e443e4c77a08450b7c74f1))
* **deps:** bump cloud.google.com/go/storage from 1.30.1 to 1.32.0 ([#103](https://github.com/rudderlabs/rudder-go-kit/issues/103)) ([cf9f48d](https://github.com/rudderlabs/rudder-go-kit/commit/cf9f48de79647937106b6cb650784a7329c31a36))
* **deps:** bump github.com/aws/aws-sdk-go from 1.44.284 to 1.45.3 ([#115](https://github.com/rudderlabs/rudder-go-kit/issues/115)) ([3dd6bd2](https://github.com/rudderlabs/rudder-go-kit/commit/3dd6bd2a0e433477f6ab32983e95bce1da8b0aa2))
* **deps:** bump github.com/go-chi/chi/v5 from 5.0.8 to 5.0.10 ([#102](https://github.com/rudderlabs/rudder-go-kit/issues/102)) ([3c3359c](https://github.com/rudderlabs/rudder-go-kit/commit/3c3359ccc80755366611cc799037a2c1936724a8))
* **deps:** bump github.com/minio/minio-go/v7 from 7.0.57 to 7.0.63 ([#106](https://github.com/rudderlabs/rudder-go-kit/issues/106)) ([4e2a6d4](https://github.com/rudderlabs/rudder-go-kit/commit/4e2a6d4880a213ff64cfeedbcdf5aa29226e3d01))
* **deps:** bump go.uber.org/zap from 1.24.0 to 1.25.0 ([#101](https://github.com/rudderlabs/rudder-go-kit/issues/101)) ([ed1ba46](https://github.com/rudderlabs/rudder-go-kit/commit/ed1ba46faede7b94e937a99eec0079b16405d5ff))
* upgrade go version 1.21 ([#116](https://github.com/rudderlabs/rudder-go-kit/issues/116)) ([1f9dde1](https://github.com/rudderlabs/rudder-go-kit/commit/1f9dde1f20b705c13fd24ff2c8244dd032054141))

## [0.15.7](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.6...v0.15.7) (2023-08-25)


### Miscellaneous

* add sqlutil.PrintRowsToTable ([#98](https://github.com/rudderlabs/rudder-go-kit/issues/98)) ([df55621](https://github.com/rudderlabs/rudder-go-kit/commit/df55621fa442969ecd59b95cc1aa13f0919a9bea))
* **deps:** bump github.com/prometheus/common from 0.42.0 to 0.44.0 ([#87](https://github.com/rudderlabs/rudder-go-kit/issues/87)) ([ad6cb2e](https://github.com/rudderlabs/rudder-go-kit/commit/ad6cb2e0035d5e4048f576cf6b9a28dcee69ce92))
* **deps:** bump github.com/shirou/gopsutil/v3 from 3.23.4 to 3.23.7 ([#85](https://github.com/rudderlabs/rudder-go-kit/issues/85)) ([94832ed](https://github.com/rudderlabs/rudder-go-kit/commit/94832ed00611e5a095b001ea40892e6555fd222d))
* **deps:** bump github.com/spf13/viper from 1.15.0 to 1.16.0 ([#94](https://github.com/rudderlabs/rudder-go-kit/issues/94)) ([067692c](https://github.com/rudderlabs/rudder-go-kit/commit/067692c1be152a4d854720981410405a56709056))
* **deps:** bump github.com/throttled/throttled/v2 from 2.11.0 to 2.12.0 ([#84](https://github.com/rudderlabs/rudder-go-kit/issues/84)) ([61a0b55](https://github.com/rudderlabs/rudder-go-kit/commit/61a0b5545fd6c6097a16a5c9fc96be7465488907))
* **deps:** bump google.golang.org/api from 0.128.0 to 0.138.0 ([#95](https://github.com/rudderlabs/rudder-go-kit/issues/95)) ([1b6a3aa](https://github.com/rudderlabs/rudder-go-kit/commit/1b6a3aa681ad8fe836996feedccb91f592657b8f))

## [0.15.6](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.5...v0.15.6) (2023-08-23)


### Bug Fixes

* gcs manager race ([#96](https://github.com/rudderlabs/rudder-go-kit/issues/96)) ([64602fd](https://github.com/rudderlabs/rudder-go-kit/commit/64602fd8c1f1b9d985a22dd1b0f8beb772440eb4))

## [0.15.5](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.4...v0.15.5) (2023-08-08)


### Bug Fixes

* s3 manager data race ([#88](https://github.com/rudderlabs/rudder-go-kit/issues/88)) ([7e2ef74](https://github.com/rudderlabs/rudder-go-kit/commit/7e2ef7471339307b8b11472f9aa1abbb88f963ef))

## [0.15.4](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.3...v0.15.4) (2023-07-26)


### Bug Fixes

* minio manager data race ([#82](https://github.com/rudderlabs/rudder-go-kit/issues/82)) ([f55c7ae](https://github.com/rudderlabs/rudder-go-kit/commit/f55c7aed4a8876426c24cf4992f14e694cc2e409))

## [0.15.3](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.2...v0.15.3) (2023-07-03)


### Miscellaneous

* make limiter's end function idempotent ([#71](https://github.com/rudderlabs/rudder-go-kit/issues/71)) ([cd5826d](https://github.com/rudderlabs/rudder-go-kit/commit/cd5826d6cd3edb8cf1067b73b63759e98ead4acc))

## [0.15.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.1...v0.15.2) (2023-07-03)


### Miscellaneous

* otel stable metric API update v1.16.0 / v0.39.0 ([#72](https://github.com/rudderlabs/rudder-go-kit/issues/72)) ([2fbe8cd](https://github.com/rudderlabs/rudder-go-kit/commit/2fbe8cda7851b070baf1971af8fe69c2b0c399c3))

## [0.15.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.15.0...v0.15.1) (2023-06-21)


### Miscellaneous

* add filemanager ([#69](https://github.com/rudderlabs/rudder-go-kit/issues/69)) ([e44d447](https://github.com/rudderlabs/rudder-go-kit/commit/e44d44728af1399bf767957982c0ff090399976d))

## [0.15.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.14.3...v0.15.0) (2023-06-08)


### Features

* redact unknown paths by default in chiware StatMiddleware ([#67](https://github.com/rudderlabs/rudder-go-kit/issues/67)) ([bb7d78d](https://github.com/rudderlabs/rudder-go-kit/commit/bb7d78dfe0cd52b4baf5de56c05a9606ee7ba7ce))

## [0.14.3](https://github.com/rudderlabs/rudder-go-kit/compare/v0.14.2...v0.14.3) (2023-05-31)


### Miscellaneous

* avoid starting a limiter using a limit less than zero ([#64](https://github.com/rudderlabs/rudder-go-kit/issues/64)) ([6b9295d](https://github.com/rudderlabs/rudder-go-kit/commit/6b9295df87fee88a98dbc5cefb2e1438f2d6943a))

## [0.14.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.14.1...v0.14.2) (2023-05-29)


### Bug Fixes

* otel prometheus duplicated attributes ([#62](https://github.com/rudderlabs/rudder-go-kit/issues/62)) ([34c9d32](https://github.com/rudderlabs/rudder-go-kit/commit/34c9d323a1e26d18298df8826a0748da43faf168))

## [0.14.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.14.0...v0.14.1) (2023-05-25)


### Bug Fixes

* limiter not respecting WithLimiterTags option ([#60](https://github.com/rudderlabs/rudder-go-kit/issues/60)) ([257e165](https://github.com/rudderlabs/rudder-go-kit/commit/257e165826a144ed7875fdb7ad86881c16dd652d))

## [0.14.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.13.5...v0.14.0) (2023-05-24)


### Features

* add bytesize, queue, mem, profiler and sync packages ([#50](https://github.com/rudderlabs/rudder-go-kit/issues/50)) ([4bfc4e1](https://github.com/rudderlabs/rudder-go-kit/commit/4bfc4e12d074d0e01e03cc4e25d05ecb14bd9587))

## [0.13.5](https://github.com/rudderlabs/rudder-go-kit/compare/v0.13.4...v0.13.5) (2023-05-23)


### Bug Fixes

* gauge continuity ([#52](https://github.com/rudderlabs/rudder-go-kit/issues/52)) ([698f566](https://github.com/rudderlabs/rudder-go-kit/commit/698f566a062f2925ee7e5925b793b443f51145dc))

## [0.13.4](https://github.com/rudderlabs/rudder-go-kit/compare/v0.13.3...v0.13.4) (2023-05-19)


### Bug Fixes

* instanceName consistency ([#51](https://github.com/rudderlabs/rudder-go-kit/issues/51)) ([b831f9b](https://github.com/rudderlabs/rudder-go-kit/commit/b831f9b971fd3b9ce4949bfded572492c309ec91))


### Miscellaneous

* use official pulsar container for arm64 ([#48](https://github.com/rudderlabs/rudder-go-kit/issues/48)) ([9eeee52](https://github.com/rudderlabs/rudder-go-kit/commit/9eeee525bed6c9dcaa25ecbd894fe689df7e49af))

## [0.13.3](https://github.com/rudderlabs/rudder-go-kit/compare/v0.13.2...v0.13.3) (2023-05-08)


### Miscellaneous

* use common pat for release please ([#46](https://github.com/rudderlabs/rudder-go-kit/issues/46)) ([10142df](https://github.com/rudderlabs/rudder-go-kit/commit/10142df2ff103fe2fb6cba77c2aaa4144b2d20eb))

## [0.13.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.13.1...v0.13.2) (2023-05-08)


### Bug Fixes

* race condition for err variable when statsd start is called ([#44](https://github.com/rudderlabs/rudder-go-kit/issues/44)) ([5702038](https://github.com/rudderlabs/rudder-go-kit/commit/5702038bfb255d4efbac3a1493a4b5fc788ff989))

## [0.13.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.13.0...v0.13.1) (2023-05-04)


### Miscellaneous

* add logs during StatsD client creation ([#43](https://github.com/rudderlabs/rudder-go-kit/issues/43)) ([915ff80](https://github.com/rudderlabs/rudder-go-kit/commit/915ff804710b7d5ca9db8f849e143abe4cab2825))
* **deps:** bump github.com/lib/pq from 1.10.8 to 1.10.9 ([#31](https://github.com/rudderlabs/rudder-go-kit/issues/31)) ([74779d8](https://github.com/rudderlabs/rudder-go-kit/commit/74779d836e28eef92e214e40c86847ec54a9540d))

## [0.13.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.12.0...v0.13.0) (2023-04-19)


### Features

* package ro - memoize ([#30](https://github.com/rudderlabs/rudder-go-kit/issues/30)) ([ea99f1c](https://github.com/rudderlabs/rudder-go-kit/commit/ea99f1c6215654efca174340b5749c7a840a647c))


### Miscellaneous

* **deps:** bump github.com/cenkalti/backoff/v4 from 4.2.0 to 4.2.1 ([#28](https://github.com/rudderlabs/rudder-go-kit/issues/28)) ([8d12fec](https://github.com/rudderlabs/rudder-go-kit/commit/8d12fec400011d6e51e572ff1383eb876f9311d3))
* **deps:** bump github.com/lib/pq from 1.10.7 to 1.10.8 ([#27](https://github.com/rudderlabs/rudder-go-kit/issues/27)) ([1886ac8](https://github.com/rudderlabs/rudder-go-kit/commit/1886ac8dd8e6f8f7212693fcc487cfce5f7cc006))
* **deps:** bump github.com/ory/dockertest/v3 from 3.9.1 to 3.10.0 ([#26](https://github.com/rudderlabs/rudder-go-kit/issues/26)) ([60bfaa1](https://github.com/rudderlabs/rudder-go-kit/commit/60bfaa1487c21c80ad49d0fdffb63729946e2b80))

## [0.12.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.11.2...v0.12.0) (2023-04-14)


### Features

* throttling ([#21](https://github.com/rudderlabs/rudder-go-kit/issues/21)) ([2423d26](https://github.com/rudderlabs/rudder-go-kit/commit/2423d2695891e32580a266fc54574e3cdd596205))


### Miscellaneous

* **deps:** bump github.com/prometheus/client_golang from 1.14.0 to 1.15.0 ([#24](https://github.com/rudderlabs/rudder-go-kit/issues/24)) ([dd6400d](https://github.com/rudderlabs/rudder-go-kit/commit/dd6400def7bc7d327b334f4c8b3e5ccfde4c5b7c))

## [0.11.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.11.1...v0.11.2) (2023-04-13)


### Miscellaneous

* remove get instance id function ([#22](https://github.com/rudderlabs/rudder-go-kit/issues/22)) ([dfd9478](https://github.com/rudderlabs/rudder-go-kit/commit/dfd9478e8b798a9862b74ed0ef4c6d6ae32b6c8b))

## [0.11.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.11.0...v0.11.1) (2023-04-11)


### Bug Fixes

* handle the case when the gateway suffix (in HA mode) ends up being a number ([#16](https://github.com/rudderlabs/rudder-go-kit/issues/16)) ([f6ee7d4](https://github.com/rudderlabs/rudder-go-kit/commit/f6ee7d4732c897bef170565e38ee5de582c81801))
* wait for pulsar proper startup ([#17](https://github.com/rudderlabs/rudder-go-kit/issues/17)) ([154ff20](https://github.com/rudderlabs/rudder-go-kit/commit/154ff20abeb1d8fe5c3d32c49452085e155ae83e))


### Miscellaneous

* **deps:** bump github.com/docker/docker from 20.10.21+incompatible to 20.10.24+incompatible ([#19](https://github.com/rudderlabs/rudder-go-kit/issues/19)) ([c4499f7](https://github.com/rudderlabs/rudder-go-kit/commit/c4499f7029927dd5493066031ecd7e115a35c1fc))
* **deps:** bump github.com/opencontainers/runc from 1.1.4 to 1.1.5 ([#18](https://github.com/rudderlabs/rudder-go-kit/issues/18)) ([6d3d4f5](https://github.com/rudderlabs/rudder-go-kit/commit/6d3d4f535b6541320d0963ef6470b6492783f0bc))

## [0.11.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.10.0...v0.11.0) (2023-03-30)


### Features

* otel metrics endpoint ([#13](https://github.com/rudderlabs/rudder-go-kit/issues/13)) ([f08036e](https://github.com/rudderlabs/rudder-go-kit/commit/f08036ef557cb52f23214383652ef31108456137))

## [0.10.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.9.0...v0.10.0) (2023-03-28)


### Features

* otel default buckets ([#10](https://github.com/rudderlabs/rudder-go-kit/issues/10)) ([ef9e011](https://github.com/rudderlabs/rudder-go-kit/commit/ef9e011b8da8cfd2ebb6f0d99f2b3ea09bb0cf71))


### Miscellaneous

* add StatMiddleware for chi router ([#14](https://github.com/rudderlabs/rudder-go-kit/issues/14)) ([40db9ff](https://github.com/rudderlabs/rudder-go-kit/commit/40db9ff2389be974e9045f998bbf9bbbf1cf3a97))

## 0.9.0 (2023-03-15)


### Features

* config logger and stats common packages ([#1](https://github.com/rudderlabs/rudder-go-kit/issues/1)) ([b036bd4](https://github.com/rudderlabs/rudder-go-kit/commit/b036bd49f2425b16b9ae300ddf27fd20cbe02267))


### Miscellaneous

* release 0.9.0 ([dd26329](https://github.com/rudderlabs/rudder-go-kit/commit/dd2632950c94940ec771a72fadd25e35bb1a5b6f))
