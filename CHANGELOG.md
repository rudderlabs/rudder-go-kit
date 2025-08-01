# Changelog

## [0.59.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.59.1...v0.59.2) (2025-07-31)


### Miscellaneous

* add sync.OnceEvery for performing actions at most once every interval ([#885](https://github.com/rudderlabs/rudder-go-kit/issues/885)) ([23dedda](https://github.com/rudderlabs/rudder-go-kit/commit/23deddafed3d47aa6763bc596c2e13c40339519b))
* **deps:** bump cloud.google.com/go/storage from 1.55.0 to 1.56.0 in the frequent group ([#880](https://github.com/rudderlabs/rudder-go-kit/issues/880)) ([2401170](https://github.com/rudderlabs/rudder-go-kit/commit/2401170fba3f04c0db08e0aa2483d1c7fb16587e))
* **deps:** bump github.com/docker/docker from 28.3.1+incompatible to 28.3.3+incompatible in the go_modules group ([#882](https://github.com/rudderlabs/rudder-go-kit/issues/882)) ([5e27c31](https://github.com/rudderlabs/rudder-go-kit/commit/5e27c311cf94012d06e873f8e1ce72f59e865d0e))
* **deps:** bump the all group with 3 updates ([#884](https://github.com/rudderlabs/rudder-go-kit/issues/884)) ([2f5696e](https://github.com/rudderlabs/rudder-go-kit/commit/2f5696e890ead601eb17b557b9a7b6ef806cdfd4))
* **deps:** bump the all group with 8 updates ([#881](https://github.com/rudderlabs/rudder-go-kit/issues/881)) ([0358a91](https://github.com/rudderlabs/rudder-go-kit/commit/0358a91d3c2f5629a2c561095c2aee779e790d53))

## [0.59.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.59.0...v0.59.1) (2025-07-25)


### Bug Fixes

* transformerarm64 docker image exits with SIGSEGV after calling a user transformation ([#875](https://github.com/rudderlabs/rudder-go-kit/issues/875)) ([b221da5](https://github.com/rudderlabs/rudder-go-kit/commit/b221da500f09a4b031831b4e18bd80b1e316ac38))


### Miscellaneous

* **deps:** bump the all group across 1 directory with 9 updates ([#873](https://github.com/rudderlabs/rudder-go-kit/issues/873)) ([a2d5085](https://github.com/rudderlabs/rudder-go-kit/commit/a2d5085aefb29cbf3e5a691d9acec7bddb3ae071))
* remove sync.ErrGroup ([#870](https://github.com/rudderlabs/rudder-go-kit/issues/870)) ([025ce8e](https://github.com/rudderlabs/rudder-go-kit/commit/025ce8ec56ce2598dff1fd1a7c7dd1de1eaf5f3d))
* use latest postgres ([#874](https://github.com/rudderlabs/rudder-go-kit/issues/874)) ([f802363](https://github.com/rudderlabs/rudder-go-kit/commit/f8023634c2e743e733b4d9f0313dc6487d340d1f))

## [0.59.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.58.0...v0.59.0) (2025-07-18)


### Features

* new int slice field ([#866](https://github.com/rudderlabs/rudder-go-kit/issues/866)) ([b166544](https://github.com/rudderlabs/rudder-go-kit/commit/b16654463be6a685ad5fd5e188bb8b34488fe1d9))


### Miscellaneous

* add get bucket region helper ([#867](https://github.com/rudderlabs/rudder-go-kit/issues/867)) ([c51fe1f](https://github.com/rudderlabs/rudder-go-kit/commit/c51fe1f4e2987d68cef7a1382b1df41090493da4))
* **deps:** bump the all group across 1 directory with 8 updates ([#865](https://github.com/rudderlabs/rudder-go-kit/issues/865)) ([217700d](https://github.com/rudderlabs/rudder-go-kit/commit/217700d9e48966d198675dacdf94f904ceb6b768))

## [0.58.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.57.1...v0.58.0) (2025-07-15)


### Features

* add default HTTP client and options for custom transport and timeout ([#854](https://github.com/rudderlabs/rudder-go-kit/issues/854)) ([a4a881c](https://github.com/rudderlabs/rudder-go-kit/commit/a4a881c6bbc091fa20b5ff21e4d1697922a2960b))

## [0.57.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.57.0...v0.57.1) (2025-07-14)


### Bug Fixes

* s3 manager v2 with prefix sanitization and update tests ([#859](https://github.com/rudderlabs/rudder-go-kit/issues/859)) ([15038b0](https://github.com/rudderlabs/rudder-go-kit/commit/15038b0974600531f975a3b5b91d56a59d103dbf))


### Miscellaneous

* parse region and remove useGlue from config ([#846](https://github.com/rudderlabs/rudder-go-kit/issues/846)) ([1b18878](https://github.com/rudderlabs/rudder-go-kit/commit/1b188787763d86dccef2289f6fd9edc364b6d748))
* skip setting session region and ignore region fetch errors ([#858](https://github.com/rudderlabs/rudder-go-kit/issues/858)) ([c931aaa](https://github.com/rudderlabs/rudder-go-kit/commit/c931aaa1c31ff4987c3e7572eaf2ca8ecfcf0885))
* update dependencies ([#853](https://github.com/rudderlabs/rudder-go-kit/issues/853)) ([671bc0f](https://github.com/rudderlabs/rudder-go-kit/commit/671bc0f7751bf9b8a73716754de90ca1ee0615ae))

## [0.57.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.56.4...v0.57.0) (2025-07-02)


### Features

* add JSON parser implementation with switchable libraries  ([#809](https://github.com/rudderlabs/rudder-go-kit/issues/809)) ([d6a8eb1](https://github.com/rudderlabs/rudder-go-kit/commit/d6a8eb1070d84bbb1ed51cdeb217f67658b1d1ba))

## [0.56.4](https://github.com/rudderlabs/rudder-go-kit/compare/v0.56.3...v0.56.4) (2025-07-01)


### Miscellaneous

* remove scylla ([#850](https://github.com/rudderlabs/rudder-go-kit/issues/850)) ([0c83c6b](https://github.com/rudderlabs/rudder-go-kit/commit/0c83c6bb7723ea33a71ac9d3ff60d28cc483d63b))

## [0.56.3](https://github.com/rudderlabs/rudder-go-kit/compare/v0.56.2...v0.56.3) (2025-07-01)


### Bug Fixes

* non-reloadable config values getting updated during change detection ([#848](https://github.com/rudderlabs/rudder-go-kit/issues/848)) ([e533c50](https://github.com/rudderlabs/rudder-go-kit/commit/e533c5043afdf1c10aa29ad26b1cf64516d3c920))


### Miscellaneous

* enforcing non-sugared logger ([#847](https://github.com/rudderlabs/rudder-go-kit/issues/847)) ([ab67d10](https://github.com/rudderlabs/rudder-go-kit/commit/ab67d101ed3a983d8d58dfd2881155497f7f1f2b))

## [0.56.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.56.1...v0.56.2) (2025-06-24)


### Miscellaneous

* add gcs option for uploading if object doesn't exist ([#843](https://github.com/rudderlabs/rudder-go-kit/issues/843)) ([be2aff1](https://github.com/rudderlabs/rudder-go-kit/commit/be2aff1a9c57721f0a33e60d74f7c280155f3825))

## [0.56.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.56.0...v0.56.1) (2025-06-23)


### Bug Fixes

* an empty endpoint is causing errors in s3 filemanager while using sdk v2 ([#838](https://github.com/rudderlabs/rudder-go-kit/issues/838)) ([7c0bc18](https://github.com/rudderlabs/rudder-go-kit/commit/7c0bc18761211b7281602975bc095bbe8a06eeb1))

## [0.56.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.55.2...v0.56.0) (2025-06-19)


### Features

* add support for S3 select objects ([#815](https://github.com/rudderlabs/rudder-go-kit/issues/815)) ([8ea56b9](https://github.com/rudderlabs/rudder-go-kit/commit/8ea56b9c6f066ecbaf5613a1bffaff8355c1b0d1))
* introduce sync.ErrGroup ([#836](https://github.com/rudderlabs/rudder-go-kit/issues/836)) ([6684771](https://github.com/rudderlabs/rudder-go-kit/commit/66847717eb75cf400560d30cf1acdb042a91fc9a))

## [0.55.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.55.1...v0.55.2) (2025-06-12)


### Bug Fixes

* improve error handling and clarify retry strategy behavior ([#831](https://github.com/rudderlabs/rudder-go-kit/issues/831)) ([733eed0](https://github.com/rudderlabs/rudder-go-kit/commit/733eed082e7d28f7dc3deac2b1200edc80bb0f05))


### Miscellaneous

* update confluent-kafka-go to v2.10.1 and go to 1.24.4 ([#830](https://github.com/rudderlabs/rudder-go-kit/issues/830)) ([dba4e28](https://github.com/rudderlabs/rudder-go-kit/commit/dba4e2838f8353bf05f17a8f5feb7a86bd5ed93d))

## [0.55.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.55.0...v0.55.1) (2025-06-11)


### Miscellaneous

* **deps:** bump the all group across 1 directory with 31 updates ([#826](https://github.com/rudderlabs/rudder-go-kit/issues/826)) ([52fb371](https://github.com/rudderlabs/rudder-go-kit/commit/52fb37101eeabca758d7d713a84bde775c753c61))

## [0.55.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.54.1...v0.55.0) (2025-06-08)


### Features

* **config:** detect config changes for reloadable and non-reloadable keys ([#817](https://github.com/rudderlabs/rudder-go-kit/issues/817)) ([c00e4dd](https://github.com/rudderlabs/rudder-go-kit/commit/c00e4ddc22dbfb094edfa84b4640b03c60d231db))

## [0.54.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.54.0...v0.54.1) (2025-06-03)


### Miscellaneous

* sonnet should use its own marshaller in jsonrs ([#821](https://github.com/rudderlabs/rudder-go-kit/issues/821)) ([6c913c5](https://github.com/rudderlabs/rudder-go-kit/commit/6c913c5c00ab0978dcdd576881fecad6e50da168))

## [0.54.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.53.1...v0.54.0) (2025-06-03)


### Features

* add a retryable HTTP client with exponential backoff ([#818](https://github.com/rudderlabs/rudder-go-kit/issues/818)) ([eea4af8](https://github.com/rudderlabs/rudder-go-kit/commit/eea4af8945d29dce7c75fd85e57efa486214af5a))

## [0.53.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.53.0...v0.53.1) (2025-06-02)


### Bug Fixes

* duplicate prefix S3 Upload ([#816](https://github.com/rudderlabs/rudder-go-kit/issues/816)) ([8f25710](https://github.com/rudderlabs/rudder-go-kit/commit/8f257108e2a1e65ea34b51a4829baf5fe1fe809f))

## [0.53.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.52.0...v0.53.0) (2025-05-28)


### Features

* introduce netutil package ([#813](https://github.com/rudderlabs/rudder-go-kit/issues/813)) ([186c027](https://github.com/rudderlabs/rudder-go-kit/commit/186c02714025787ffed16e05152a5d8eafcff2a4))

## [0.52.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.51.0...v0.52.0) (2025-05-23)


### Features

* introduce aws sdk v2 ([#795](https://github.com/rudderlabs/rudder-go-kit/issues/795)) ([d4f763a](https://github.com/rudderlabs/rudder-go-kit/commit/d4f763ac78941058562baa1ba2e96f1c14cdf03e))

## [0.51.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.50.2...v0.51.0) (2025-05-19)


### Features

* implement JSON switching mechanism with multiple libraries ([#803](https://github.com/rudderlabs/rudder-go-kit/issues/803)) ([906400e](https://github.com/rudderlabs/rudder-go-kit/commit/906400e6f09adf54e7557c2f19291009392f418b))

## [0.50.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.50.1...v0.50.2) (2025-05-08)


### Miscellaneous

* allow transformer container to bind to specific IP ([#800](https://github.com/rudderlabs/rudder-go-kit/issues/800)) ([57ff124](https://github.com/rudderlabs/rudder-go-kit/commit/57ff124cd734a0dce82c1d2d48e188c495b33308))

## [0.50.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.50.0...v0.50.1) (2025-05-08)


### Bug Fixes

* no host.docker.internal for linux ([#797](https://github.com/rudderlabs/rudder-go-kit/issues/797)) ([5089859](https://github.com/rudderlabs/rudder-go-kit/commit/5089859986861801bdbd6ea0e387689d2f56ee58))


### Miscellaneous

* **deps:** bump github.com/golang-jwt/jwt/v5 from 5.2.1 to 5.2.2 in the go_modules group ([#799](https://github.com/rudderlabs/rudder-go-kit/issues/799)) ([a3ac44d](https://github.com/rudderlabs/rudder-go-kit/commit/a3ac44d48371ba4dbc7eb64f554132ec6687c00e))

## [0.50.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.49.1...v0.50.0) (2025-05-05)


### Features

* add MapLookup function ([#790](https://github.com/rudderlabs/rudder-go-kit/issues/790)) ([6d9eb7e](https://github.com/rudderlabs/rudder-go-kit/commit/6d9eb7e8b6a8001720477041c3ce81638bae3eae))

## [0.49.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.49.0...v0.49.1) (2025-04-29)


### Miscellaneous

* **deps:** bump gitleaks/gitleaks-action from 2.3.7 to 2.3.9 ([#782](https://github.com/rudderlabs/rudder-go-kit/issues/782)) ([4ed7686](https://github.com/rudderlabs/rudder-go-kit/commit/4ed7686b815af48a559fe81fb48fcf89ba362ce1))
* **deps:** bump google.golang.org/api from 0.224.0 to 0.228.0 in the frequent group ([#765](https://github.com/rudderlabs/rudder-go-kit/issues/765)) ([ca3a1d6](https://github.com/rudderlabs/rudder-go-kit/commit/ca3a1d6663d20dd470abc85d03a33f45b4eb99e7))
* expose kafka producer batch bytes configuration property ([#788](https://github.com/rudderlabs/rudder-go-kit/issues/788)) ([e06ee56](https://github.com/rudderlabs/rudder-go-kit/commit/e06ee5666aca2d883739c3bca07ef051ef2f8c55))

## [0.49.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.48.3...v0.49.0) (2025-04-16)


### Features

* add trie data structure for prefix string matching ([#775](https://github.com/rudderlabs/rudder-go-kit/issues/775)) ([f00fc28](https://github.com/rudderlabs/rudder-go-kit/commit/f00fc284cbe1577ad0c928259e23937cb414e0d3))
* file manager direct uploads/downloads i.e. skip disk ([#777](https://github.com/rudderlabs/rudder-go-kit/issues/777)) ([82c1694](https://github.com/rudderlabs/rudder-go-kit/commit/82c1694aa34d20e4fbfd823139009c4eed66ddcd))
* support reloadable limiter limits ([#780](https://github.com/rudderlabs/rudder-go-kit/issues/780)) ([103a934](https://github.com/rudderlabs/rudder-go-kit/commit/103a93420fc85b348c237e892fb0343efd763a4f))


### Miscellaneous

* **deps:** bump the all group across 1 directory with 13 updates ([#776](https://github.com/rudderlabs/rudder-go-kit/issues/776)) ([309c68e](https://github.com/rudderlabs/rudder-go-kit/commit/309c68ec00b658409cb2e4e5261c47ad1e0d642c))

## [0.48.3](https://github.com/rudderlabs/rudder-go-kit/compare/v0.48.2...v0.48.3) (2025-04-15)


### Miscellaneous

* add options to run image using docker credentials ([#774](https://github.com/rudderlabs/rudder-go-kit/issues/774)) ([4b45004](https://github.com/rudderlabs/rudder-go-kit/commit/4b45004347044f0401073e122ae397449e18f3c2))
* **deps:** bump beatlabs/delete-old-branches-action from 0.0.10 to 0.0.11 ([#756](https://github.com/rudderlabs/rudder-go-kit/issues/756)) ([774086e](https://github.com/rudderlabs/rudder-go-kit/commit/774086ee161cee02691fa57f3a2e3310bcf2443b))

## [0.48.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.48.1...v0.48.2) (2025-04-02)


### Miscellaneous

* nop tracer ([#769](https://github.com/rudderlabs/rudder-go-kit/issues/769)) ([fce5502](https://github.com/rudderlabs/rudder-go-kit/commit/fce55029c7c1087f0568ba4628a0b878052a1a99))

## [0.48.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.48.0...v0.48.1) (2025-03-27)


### Miscellaneous

* aws util max idle connections property ([#763](https://github.com/rudderlabs/rudder-go-kit/issues/763)) ([e24b01c](https://github.com/rudderlabs/rudder-go-kit/commit/e24b01c6209e6427615eb819ddd7c8f16eef1b9e))
* **deps:** bump the all group across 1 directory with 25 updates ([#753](https://github.com/rudderlabs/rudder-go-kit/issues/753)) ([399cfb3](https://github.com/rudderlabs/rudder-go-kit/commit/399cfb3cb5b3b602b7a64cb7d195cf2676b365ea))

## [0.48.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.47.1...v0.48.0) (2025-03-27)


### Features

* include rss memory in calculated mem stats ([#760](https://github.com/rudderlabs/rudder-go-kit/issues/760)) ([e335ca0](https://github.com/rudderlabs/rudder-go-kit/commit/e335ca090fb8bc7f7c70a79f1bbd17ee3c776363))


### Miscellaneous

* **deps:** bump golangci/golangci-lint-action from 6 to 7 ([#761](https://github.com/rudderlabs/rudder-go-kit/issues/761)) ([f968432](https://github.com/rudderlabs/rudder-go-kit/commit/f9684327239aeb6b6a27c03591e9f42ec6c3cfd3))

## [0.47.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.47.0...v0.47.1) (2025-03-21)


### Bug Fixes

* ip not getting assigned immediately ([#757](https://github.com/rudderlabs/rudder-go-kit/issues/757)) ([e1f222e](https://github.com/rudderlabs/rudder-go-kit/commit/e1f222e8e58fe5ebfb7fa27c00f23eb23afb227b))

## [0.47.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.46.2...v0.47.0) (2025-03-19)


### Features

* allow specifying bind ip for postgres ([#742](https://github.com/rudderlabs/rudder-go-kit/issues/742)) ([fc26a01](https://github.com/rudderlabs/rudder-go-kit/commit/fc26a01541cb2077d371772197ef05e2caa58661))


### Bug Fixes

* **stats:** deadlock during stats stop ([#739](https://github.com/rudderlabs/rudder-go-kit/issues/739)) ([17ec61c](https://github.com/rudderlabs/rudder-go-kit/commit/17ec61c3b93ab24a0f4b5e90ceb5c0416faab489))

## [0.46.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.46.1...v0.46.2) (2025-01-29)


### Miscellaneous

* **deps:** bump the all group across 1 directory with 17 updates ([#731](https://github.com/rudderlabs/rudder-go-kit/issues/731)) ([8b0de4f](https://github.com/rudderlabs/rudder-go-kit/commit/8b0de4f4f2117a4cc9677cfde1e3d55d420f1dd4))
* throttling tagger ([#730](https://github.com/rudderlabs/rudder-go-kit/issues/730)) ([26e7749](https://github.com/rudderlabs/rudder-go-kit/commit/26e7749dba321324fa833f40fb59070cee0403e5))

## [0.46.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.46.0...v0.46.1) (2025-01-15)


### Miscellaneous

* **deps:** bump the all group with 4 updates ([#716](https://github.com/rudderlabs/rudder-go-kit/issues/716)) ([0e63f29](https://github.com/rudderlabs/rudder-go-kit/commit/0e63f297812c58df1b0b401a1e6a7afec779b820))
* logger with config ([#717](https://github.com/rudderlabs/rudder-go-kit/issues/717)) ([d5c4932](https://github.com/rudderlabs/rudder-go-kit/commit/d5c4932fc247dc1c444f7b874e5e71c4392a0add))

## [0.46.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.45.1...v0.46.0) (2025-01-14)


### Features

* max procs ([#714](https://github.com/rudderlabs/rudder-go-kit/issues/714)) ([5210b20](https://github.com/rudderlabs/rudder-go-kit/commit/5210b208fd38fa3fbd4fae2ee0afa82cf432b83b))


### Miscellaneous

* **deps:** bump golang.org/x/crypto from 0.31.0 to 0.32.0 in the go_modules group across 1 directory ([#709](https://github.com/rudderlabs/rudder-go-kit/issues/709)) ([d794c02](https://github.com/rudderlabs/rudder-go-kit/commit/d794c02cfc4b5c66969dc97f6b0aef4fb458a31a))
* **deps:** bump the all group across 1 directory with 20 updates ([#707](https://github.com/rudderlabs/rudder-go-kit/issues/707)) ([d4e77e9](https://github.com/rudderlabs/rudder-go-kit/commit/d4e77e9789b4e3ae8cc6ad37251642c8369770b3))

## [0.45.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.45.0...v0.45.1) (2025-01-06)


### Miscellaneous

* **deps:** bump codecov/codecov-action from 4 to 5 ([#689](https://github.com/rudderlabs/rudder-go-kit/issues/689)) ([92117f7](https://github.com/rudderlabs/rudder-go-kit/commit/92117f78c0b8ffad3c316fd20e9bd01b1254e8da))
* **deps:** bump the all group across 1 directory with 28 updates ([#692](https://github.com/rudderlabs/rudder-go-kit/issues/692)) ([bf2c5fb](https://github.com/rudderlabs/rudder-go-kit/commit/bf2c5fbad93024398526354b203ea70fc7bd1a9a))
* **deps:** bump the all group with 2 updates ([#695](https://github.com/rudderlabs/rudder-go-kit/issues/695)) ([9f5f561](https://github.com/rudderlabs/rudder-go-kit/commit/9f5f5613b4808eca37f9419c2d6ea7d37382512f))
* update pulsar endpoint to use 127.0.0.1 ([#705](https://github.com/rudderlabs/rudder-go-kit/issues/705)) ([e469555](https://github.com/rudderlabs/rudder-go-kit/commit/e4695556fa663c02573927543e5cd01198e3f58e))

## [0.45.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.44.0...v0.45.0) (2024-10-23)


### Features

* request to json function usually used as pre-step for storing r… ([#668](https://github.com/rudderlabs/rudder-go-kit/issues/668)) ([50d8320](https://github.com/rudderlabs/rudder-go-kit/commit/50d8320aaac34384b905fab4b99d7de6b04fe6d8))


### Miscellaneous

* compression benchmarks ([#678](https://github.com/rudderlabs/rudder-go-kit/issues/678)) ([21abca6](https://github.com/rudderlabs/rudder-go-kit/commit/21abca63c3bcb4cad698c815e1a8a80650c7e934))

## [0.44.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.43.0...v0.44.0) (2024-10-17)


### Features

* allow retrieval of commits for tags ([#675](https://github.com/rudderlabs/rudder-go-kit/issues/675)) ([801d27b](https://github.com/rudderlabs/rudder-go-kit/commit/801d27b091fd6b834c54bc294e70bf280eb24433))


### Miscellaneous

* add support for WithEnv and WithRepository for transformer resource ([#673](https://github.com/rudderlabs/rudder-go-kit/issues/673)) ([f907f10](https://github.com/rudderlabs/rudder-go-kit/commit/f907f1082e1246b2210be24010f8aa1d47b16d47))

## [0.43.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.42.2...v0.43.0) (2024-09-27)


### Features

* introduce db stats collector ([#648](https://github.com/rudderlabs/rudder-go-kit/issues/648)) ([8ac166b](https://github.com/rudderlabs/rudder-go-kit/commit/8ac166b620b576da4fcfe88bf5c1cda07c3ba079))


### Bug Fixes

* flaky test: slice comparison to ignore element order ([#661](https://github.com/rudderlabs/rudder-go-kit/issues/661)) ([545da59](https://github.com/rudderlabs/rudder-go-kit/commit/545da59ca5f90dd9c5f59c2df94a01fecbca6237))


### Miscellaneous

* **deps:** bump the all group across 1 directory with 2 updates ([#659](https://github.com/rudderlabs/rudder-go-kit/issues/659)) ([38eaf4a](https://github.com/rudderlabs/rudder-go-kit/commit/38eaf4ae4094f389588b19c3aa50a605bf0b66de))
* **deps:** bump the all group across 1 directory with 5 updates ([#653](https://github.com/rudderlabs/rudder-go-kit/issues/653)) ([4f65bca](https://github.com/rudderlabs/rudder-go-kit/commit/4f65bcabdfeee6022dd6263243cf9f56dfa6e906))
* scylla no ipv6 ([#658](https://github.com/rudderlabs/rudder-go-kit/issues/658)) ([d8083ae](https://github.com/rudderlabs/rudder-go-kit/commit/d8083ae8f5514abe6416bf6b4a059989e21c6752))
* transformer network ([#656](https://github.com/rudderlabs/rudder-go-kit/issues/656)) ([96fb376](https://github.com/rudderlabs/rudder-go-kit/commit/96fb376ba70d60ad9c64d55a5affa2db6d89601b))
* update go and linter version ([#660](https://github.com/rudderlabs/rudder-go-kit/issues/660)) ([1d7c2ce](https://github.com/rudderlabs/rudder-go-kit/commit/1d7c2ce98c996dd1f51fd45524b8f46a0a613cea))

## [0.42.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.42.1...v0.42.2) (2024-09-23)


### Miscellaneous

* allow overriding the bind ip address in docker containers ([#647](https://github.com/rudderlabs/rudder-go-kit/issues/647)) ([6bb5880](https://github.com/rudderlabs/rudder-go-kit/commit/6bb58808d23ced907874e9cc436dcb4822f9968f))

## [0.42.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.42.0...v0.42.1) (2024-09-18)


### Miscellaneous

* cachettl configuration option for not refreshing ttl ([#643](https://github.com/rudderlabs/rudder-go-kit/issues/643)) ([1961d98](https://github.com/rudderlabs/rudder-go-kit/commit/1961d9894a68dded99840c4dd398578be503f2e4))

## [0.42.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.41.1...v0.42.0) (2024-09-11)


### Features

* add encryption support ([#627](https://github.com/rudderlabs/rudder-go-kit/issues/627)) ([e26c3e4](https://github.com/rudderlabs/rudder-go-kit/commit/e26c3e4f67419e8d245e59ab3360f0e35002eb16))


### Miscellaneous

* **deps:** bump the all group across 1 directory with 17 updates ([#640](https://github.com/rudderlabs/rudder-go-kit/issues/640)) ([2d220c1](https://github.com/rudderlabs/rudder-go-kit/commit/2d220c1e4fadbff702f5e0b1504863e5ba9bd680))

## [0.41.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.41.0...v0.41.1) (2024-09-10)


### Bug Fixes

* gittest server `tls: unknown certificate authority` ([#638](https://github.com/rudderlabs/rudder-go-kit/issues/638)) ([bea4330](https://github.com/rudderlabs/rudder-go-kit/commit/bea4330070559ef53d6b41826465c4f7f95141b2))

## [0.41.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.40.0...v0.41.0) (2024-09-10)


### Features

* method for fetching latest commit from gittest server ([#636](https://github.com/rudderlabs/rudder-go-kit/issues/636)) ([7d4f518](https://github.com/rudderlabs/rudder-go-kit/commit/7d4f518e7883d9ceb7b32c10e2d8b85d7c71137a))

## [0.40.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.39.2...v0.40.0) (2024-09-06)


### Features

* [PRO-3387] minio enhancements ([#630](https://github.com/rudderlabs/rudder-go-kit/issues/630)) ([8a4f24a](https://github.com/rudderlabs/rudder-go-kit/commit/8a4f24a4556e51041fd8851277778e504f047f79))
* introduce gittest ([#631](https://github.com/rudderlabs/rudder-go-kit/issues/631)) ([c7d3c82](https://github.com/rudderlabs/rudder-go-kit/commit/c7d3c82edbe514c8a138e0a81cf9018fd4fb3681))


### Bug Fixes

* scylla test fails with no connections were made when creating the session ([#634](https://github.com/rudderlabs/rudder-go-kit/issues/634)) ([5e3b759](https://github.com/rudderlabs/rudder-go-kit/commit/5e3b7596e16e01cee80d0bf9efefe137dcf9a35d))

## [0.39.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.39.1...v0.39.2) (2024-09-04)


### Miscellaneous

* **deps:** bump github.com/opencontainers/runc from 1.1.13 to 1.1.14 ([#625](https://github.com/rudderlabs/rudder-go-kit/issues/625)) ([a7649f6](https://github.com/rudderlabs/rudder-go-kit/commit/a7649f6fc78698868505eaef86b9d82cb04d6bae))
* minio network ([#628](https://github.com/rudderlabs/rudder-go-kit/issues/628)) ([7b26236](https://github.com/rudderlabs/rudder-go-kit/commit/7b26236e001ae49ea647cf0a06864af24c825203))
* use gitleaks for secret scanning ([#610](https://github.com/rudderlabs/rudder-go-kit/issues/610)) ([a8ac9a5](https://github.com/rudderlabs/rudder-go-kit/commit/a8ac9a5e5b6ef97125e277c6530a08e946b3c74d))

## [0.39.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.39.0...v0.39.1) (2024-09-03)


### Bug Fixes

* compress ([#621](https://github.com/rudderlabs/rudder-go-kit/issues/621)) ([ae791c9](https://github.com/rudderlabs/rudder-go-kit/commit/ae791c9c8bd606c8758f8dcc57d81b3d875b3f90))

## [0.39.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.38.2...v0.39.0) (2024-09-03)


### Features

* compress zstd cgo ([#618](https://github.com/rudderlabs/rudder-go-kit/issues/618)) ([2402c8e](https://github.com/rudderlabs/rudder-go-kit/commit/2402c8e914be26cc8ae4e21c1e6de3df2a970cea))

## [0.38.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.38.1...v0.38.2) (2024-08-30)


### Bug Fixes

* docker containers listening on ipv6 ([#616](https://github.com/rudderlabs/rudder-go-kit/issues/616)) ([76a6758](https://github.com/rudderlabs/rudder-go-kit/commit/76a67587b145c38dee21db4461b9e624f03cba44))

## [0.38.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.38.0...v0.38.1) (2024-08-30)


### Bug Fixes

* docker containers don't listen to host.docker.internal ([#614](https://github.com/rudderlabs/rudder-go-kit/issues/614)) ([7df95c9](https://github.com/rudderlabs/rudder-go-kit/commit/7df95c9eb390b5830b1719483cfe6b68be629589))


### Miscellaneous

* **deps:** bump the all group across 1 directory with 11 updates ([#609](https://github.com/rudderlabs/rudder-go-kit/issues/609)) ([00f6635](https://github.com/rudderlabs/rudder-go-kit/commit/00f66350d0e4aa2b3a9ba52c4733f405a1b2a93c))

## [0.38.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.37.2...v0.38.0) (2024-08-27)


### Features

* compress ([#601](https://github.com/rudderlabs/rudder-go-kit/issues/601)) ([cda0a40](https://github.com/rudderlabs/rudder-go-kit/commit/cda0a40d46f3e81ce682356d7f42a34abdf95162))

## [0.37.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.37.1...v0.37.2) (2024-08-27)


### Miscellaneous

* isTruncated should be a debug log ([#606](https://github.com/rudderlabs/rudder-go-kit/issues/606)) ([075f014](https://github.com/rudderlabs/rudder-go-kit/commit/075f014068c93a1d12e27235aeda12cd58b93a37))

## [0.37.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.37.0...v0.37.1) (2024-08-26)


### Miscellaneous

* add more configuration options for creating aws sessions ([#602](https://github.com/rudderlabs/rudder-go-kit/issues/602)) ([3393b43](https://github.com/rudderlabs/rudder-go-kit/commit/3393b431a12159a20531b888672a006ae8e1f015))
* **deps:** bump google.golang.org/api from 0.193.0 to 0.194.0 in the frequent group ([#600](https://github.com/rudderlabs/rudder-go-kit/issues/600)) ([8d5a1a8](https://github.com/rudderlabs/rudder-go-kit/commit/8d5a1a839fa8ec945b8dc388501a34fa585f38f4))

## [0.37.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.36.2...v0.37.0) (2024-08-22)


### Features

* specify port binding for docker ([#592](https://github.com/rudderlabs/rudder-go-kit/issues/592)) ([ff8a8e2](https://github.com/rudderlabs/rudder-go-kit/commit/ff8a8e20baf05f73642e484567bc07d61db251ac))


### Miscellaneous

* **deps:** bump github.com/docker/docker from 25.0.5+incompatible to 25.0.6+incompatible ([#582](https://github.com/rudderlabs/rudder-go-kit/issues/582)) ([9f5c19f](https://github.com/rudderlabs/rudder-go-kit/commit/9f5c19f0076cf11b521ffef03a54832e61c2e718))
* **deps:** bump the all group across 1 directory with 25 updates ([#590](https://github.com/rudderlabs/rudder-go-kit/issues/590)) ([c31d397](https://github.com/rudderlabs/rudder-go-kit/commit/c31d397aa7256afafeb0b453a3b06fa2f971b1d4))
* **deps:** bump the all group across 1 directory with 3 updates ([#599](https://github.com/rudderlabs/rudder-go-kit/issues/599)) ([5e45a8a](https://github.com/rudderlabs/rudder-go-kit/commit/5e45a8a229a3ef265b484685ddb6903b90e7360b))

## [0.36.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.36.1...v0.36.2) (2024-08-13)


### Miscellaneous

* jit secrets ([#589](https://github.com/rudderlabs/rudder-go-kit/issues/589)) ([b53a005](https://github.com/rudderlabs/rudder-go-kit/commit/b53a00589b7915bc307f15e7792d86d88b5f1392))

## [0.36.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.36.0...v0.36.1) (2024-08-01)


### Miscellaneous

* improve scylla container health check ([#580](https://github.com/rudderlabs/rudder-go-kit/issues/580)) ([7428dce](https://github.com/rudderlabs/rudder-go-kit/commit/7428dcedf4c7480ca982de8643b7c3af642cbba2))

## [0.36.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.35.1...v0.36.0) (2024-07-30)


### Features

* create scylla resource ([#576](https://github.com/rudderlabs/rudder-go-kit/issues/576)) ([461d676](https://github.com/rudderlabs/rudder-go-kit/commit/461d676c3b86afef3c8742975602a4bb78ef017c))


### Miscellaneous

* **deps:** bump the frequent group across 1 directory with 3 updates ([#569](https://github.com/rudderlabs/rudder-go-kit/issues/569)) ([b7e92b3](https://github.com/rudderlabs/rudder-go-kit/commit/b7e92b39e2ada75462cc21b80eeb48b910b252cd))
* disable telemetry in gcs client ([#575](https://github.com/rudderlabs/rudder-go-kit/issues/575)) ([3ce22bf](https://github.com/rudderlabs/rudder-go-kit/commit/3ce22bfaf90aa89faa53414e86cd1dfabae4fc66))

## [0.35.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.35.0...v0.35.1) (2024-07-23)


### Miscellaneous

* **deps:** bump google.golang.org/grpc from 1.64.0 to 1.64.1 ([#556](https://github.com/rudderlabs/rudder-go-kit/issues/556)) ([8724b87](https://github.com/rudderlabs/rudder-go-kit/commit/8724b8716f16d113486edae4563d05d24c90358d))
* sftp filemanager mocks ([#567](https://github.com/rudderlabs/rudder-go-kit/issues/567)) ([f2c340a](https://github.com/rudderlabs/rudder-go-kit/commit/f2c340aa076b564c60c8a7670284bd9051297694))

## [0.35.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.34.2...v0.35.0) (2024-07-22)


### Features

* minio testhelper add content method ([#524](https://github.com/rudderlabs/rudder-go-kit/issues/524)) ([201797b](https://github.com/rudderlabs/rudder-go-kit/commit/201797bafa240053940fbf6402cb79c1d385f253))


### Miscellaneous

* **deps:** bump the frequent group across 1 directory with 2 updates ([#519](https://github.com/rudderlabs/rudder-go-kit/issues/519)) ([724dc2f](https://github.com/rudderlabs/rudder-go-kit/commit/724dc2f03c491bd71229b31fb0a4b409091998e1))
* move to gomock uber ([#564](https://github.com/rudderlabs/rudder-go-kit/issues/564)) ([0ef84cf](https://github.com/rudderlabs/rudder-go-kit/commit/0ef84cf536b1a0821393d7676d22334c74328dbf))

## [0.34.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.34.1...v0.34.2) (2024-06-05)


### Bug Fixes

* memstats new stat should reuse measurement ([#516](https://github.com/rudderlabs/rudder-go-kit/issues/516)) ([ab7c8f6](https://github.com/rudderlabs/rudder-go-kit/commit/ab7c8f63ea647e470999d7129a2aee021f221e4e))


### Miscellaneous

* **deps:** bump actions/checkout from 2 to 4 ([#506](https://github.com/rudderlabs/rudder-go-kit/issues/506)) ([e3525a9](https://github.com/rudderlabs/rudder-go-kit/commit/e3525a9ca851957ff57042f48573df54459821ed))
* **deps:** bump actions/setup-go from 3 to 5 ([#508](https://github.com/rudderlabs/rudder-go-kit/issues/508)) ([60c253e](https://github.com/rudderlabs/rudder-go-kit/commit/60c253ef7e0ce6ff44c09a0c95f63ac6efc523b9))
* **deps:** bump actions/stale from 5 to 9 ([#507](https://github.com/rudderlabs/rudder-go-kit/issues/507)) ([4af5318](https://github.com/rudderlabs/rudder-go-kit/commit/4af5318db573f36109682d85180103c5f14a7bf2))
* **deps:** bump codecov/codecov-action from 3 to 4 ([#509](https://github.com/rudderlabs/rudder-go-kit/issues/509)) ([e726f0c](https://github.com/rudderlabs/rudder-go-kit/commit/e726f0c1d0bbde37a048d577946be8395ba8a34e))
* **deps:** bump github.com/aws/aws-sdk-go from 1.53.12 to 1.53.15 in the frequent group across 1 directory ([#515](https://github.com/rudderlabs/rudder-go-kit/issues/515)) ([1441aae](https://github.com/rudderlabs/rudder-go-kit/commit/1441aae765f998240decfcba834bae4eed44d8b6))
* **deps:** bump the all group with 3 updates ([#513](https://github.com/rudderlabs/rudder-go-kit/issues/513)) ([d6b4ae1](https://github.com/rudderlabs/rudder-go-kit/commit/d6b4ae199c45bf37b1967e8b3e7377ff7a62ad9c))
* upgrade to go1.22.4 and standard libraries ([#517](https://github.com/rudderlabs/rudder-go-kit/issues/517)) ([3283a98](https://github.com/rudderlabs/rudder-go-kit/commit/3283a98f94b2a9579af5f2ca364c778b0970e343))

## [0.34.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.34.0...v0.34.1) (2024-05-30)


### Miscellaneous

* **ci:** general ci and tooling improvement ([#498](https://github.com/rudderlabs/rudder-go-kit/issues/498)) ([1b83f9a](https://github.com/rudderlabs/rudder-go-kit/commit/1b83f9aeea9b5bd7923cd0d1fb6a2fe5cbc5ceee))
* **deps:** bump actions/labeler from 4 to 5 ([#501](https://github.com/rudderlabs/rudder-go-kit/issues/501)) ([7647693](https://github.com/rudderlabs/rudder-go-kit/commit/76476934d5e6b686db31a7f0c32c71f13e46e436))
* **deps:** bump amannn/action-semantic-pull-request from 4 to 5 ([#500](https://github.com/rudderlabs/rudder-go-kit/issues/500)) ([2c1ff5f](https://github.com/rudderlabs/rudder-go-kit/commit/2c1ff5fa735828192ab74f9865bfebf969675b91))
* **deps:** bump beatlabs/delete-old-branches-action from 0.0.9 to 0.0.10 ([#499](https://github.com/rudderlabs/rudder-go-kit/issues/499)) ([86e0dbb](https://github.com/rudderlabs/rudder-go-kit/commit/86e0dbb79399f0c8404af913ff407c850dc3467c))
* **deps:** bump github.com/aws/aws-sdk-go from 1.53.10 to 1.53.11 ([#495](https://github.com/rudderlabs/rudder-go-kit/issues/495)) ([0af3781](https://github.com/rudderlabs/rudder-go-kit/commit/0af37819d58054344948397cb868b7703e490457))
* **deps:** bump github.com/aws/aws-sdk-go from 1.53.11 to 1.53.12 in the frequent group ([#504](https://github.com/rudderlabs/rudder-go-kit/issues/504)) ([65125f9](https://github.com/rudderlabs/rudder-go-kit/commit/65125f96cf453b13379c9b4c0e251f1013fe2014))
* **deps:** bump github.com/confluentinc/confluent-kafka-go/v2 from 2.3.0 to 2.4.0 ([#494](https://github.com/rudderlabs/rudder-go-kit/issues/494)) ([712e34e](https://github.com/rudderlabs/rudder-go-kit/commit/712e34e6b1ccbf10b2e58e9cffca9ab7cd731f76))
* **deps:** bump github.com/docker/docker from 25.0.3+incompatible to 25.0.5+incompatible ([#497](https://github.com/rudderlabs/rudder-go-kit/issues/497)) ([3c565fc](https://github.com/rudderlabs/rudder-go-kit/commit/3c565fcd55fc2924dc096193f715840fd7de2d98))
* **deps:** bump golangci/golangci-lint-action from 3 to 6 ([#503](https://github.com/rudderlabs/rudder-go-kit/issues/503)) ([4a583a0](https://github.com/rudderlabs/rudder-go-kit/commit/4a583a05b8fb930f3cd848edb9f1573879ae7797))
* **deps:** bump the all group with 2 updates ([#505](https://github.com/rudderlabs/rudder-go-kit/issues/505)) ([284005c](https://github.com/rudderlabs/rudder-go-kit/commit/284005c73590451159f03813e4f566a491ea475e))
* **deps:** bump the opentelemetry group with 9 updates ([#493](https://github.com/rudderlabs/rudder-go-kit/issues/493)) ([d823765](https://github.com/rudderlabs/rudder-go-kit/commit/d823765abd3df92c018f7026c53a2f97eafa4997))

## [0.34.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.33.0...v0.34.0) (2024-05-29)


### Features

* introduce a ttl cache for resources ([#482](https://github.com/rudderlabs/rudder-go-kit/issues/482)) ([2d5c6b2](https://github.com/rudderlabs/rudder-go-kit/commit/2d5c6b2d86849bf78faf20bce30cf5d07fcf11f7))

## [0.33.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.32.2...v0.33.0) (2024-05-29)


### Features

* utf8 sanitizer ([#485](https://github.com/rudderlabs/rudder-go-kit/issues/485)) ([0ccb5aa](https://github.com/rudderlabs/rudder-go-kit/commit/0ccb5aaa3328cec0a22f752c1bca88dbb6451ccc))


### Miscellaneous

* **deps:** bump cloud.google.com/go/storage from 1.40.0 to 1.41.0 ([#490](https://github.com/rudderlabs/rudder-go-kit/issues/490)) ([3a4a138](https://github.com/rudderlabs/rudder-go-kit/commit/3a4a138cfc4e182756aee52c138c086d13b988b8))
* **deps:** bump github.com/aws/aws-sdk-go from 1.52.0 to 1.53.10 ([#484](https://github.com/rudderlabs/rudder-go-kit/issues/484)) ([12400b4](https://github.com/rudderlabs/rudder-go-kit/commit/12400b41c7ec2962800f631e58fe55a2041ef5fb))
* **deps:** bump github.com/prometheus/client_golang from 1.19.0 to 1.19.1 ([#470](https://github.com/rudderlabs/rudder-go-kit/issues/470)) ([c0d0145](https://github.com/rudderlabs/rudder-go-kit/commit/c0d0145feec1cb2ca7b721e0738677229f9ba38b))
* **deps:** bump google.golang.org/api from 0.177.0 to 0.182.0 ([#489](https://github.com/rudderlabs/rudder-go-kit/issues/489)) ([dcca9a7](https://github.com/rudderlabs/rudder-go-kit/commit/dcca9a7eb68303407192263b55811b0118f12b9e))
* **deps:** downgrading google.golang.org/grpc from v1.64.0 to v1.63.2 ([3a6465f](https://github.com/rudderlabs/rudder-go-kit/commit/3a6465ffae8a212400a0be12cebc7c0e7239afd2))
* expose otel version from stats library ([#488](https://github.com/rudderlabs/rudder-go-kit/issues/488)) ([6acb052](https://github.com/rudderlabs/rudder-go-kit/commit/6acb0520f2402a24b316b5b76c8bf12cbcb20cc9))
* fix etcd deprecated warning ([#491](https://github.com/rudderlabs/rudder-go-kit/issues/491)) ([757b796](https://github.com/rudderlabs/rudder-go-kit/commit/757b796dc0b3b1f3d42dbc719ce237af36b9ed1c))

## [0.32.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.32.1...v0.32.2) (2024-05-28)


### Miscellaneous

* gcs manager upload condition - DoesNotExist ([#479](https://github.com/rudderlabs/rudder-go-kit/issues/479)) ([cc87fd5](https://github.com/rudderlabs/rudder-go-kit/commit/cc87fd5f1d70d2d5afc00f7cb4c7dc225f554447))

## [0.32.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.32.0...v0.32.1) (2024-05-21)


### Miscellaneous

* optimise function config.ConfigKeyToEnv ([#475](https://github.com/rudderlabs/rudder-go-kit/issues/475)) ([0cbe3e2](https://github.com/rudderlabs/rudder-go-kit/commit/0cbe3e235747c1137436e912875da7b7882fc7c4))

## [0.32.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.31.0...v0.32.0) (2024-05-20)


### Features

* custom network for etcd and pulsar ([#468](https://github.com/rudderlabs/rudder-go-kit/issues/468)) ([71f4ea7](https://github.com/rudderlabs/rudder-go-kit/commit/71f4ea7644875b74d0e1b9c761bbf038b7fd8037))


### Bug Fixes

* sftp retry on connection lost ([#465](https://github.com/rudderlabs/rudder-go-kit/issues/465)) ([8383ee7](https://github.com/rudderlabs/rudder-go-kit/commit/8383ee768a275b592f2cd97692f0ff22a6ea2fce))


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.51.20 to 1.52.0 ([#461](https://github.com/rudderlabs/rudder-go-kit/issues/461)) ([7109f26](https://github.com/rudderlabs/rudder-go-kit/commit/7109f26594631fc0e0b3a193636387999d261dd7))
* **deps:** bump github.com/linkedin/goavro/v2 from 2.12.0 to 2.13.0 ([#464](https://github.com/rudderlabs/rudder-go-kit/issues/464)) ([5db545c](https://github.com/rudderlabs/rudder-go-kit/commit/5db545c46e10a47506d2c97c61e18b83fb6adca9))
* **deps:** bump github.com/minio/minio-go/v7 from 7.0.69 to 7.0.70 ([#462](https://github.com/rudderlabs/rudder-go-kit/issues/462)) ([bc95a3f](https://github.com/rudderlabs/rudder-go-kit/commit/bc95a3f74da38d00a2b937c4bd8a078f26931e7d))
* **deps:** bump github.com/shirou/gopsutil/v3 from 3.24.3 to 3.24.4 ([#463](https://github.com/rudderlabs/rudder-go-kit/issues/463)) ([a99434b](https://github.com/rudderlabs/rudder-go-kit/commit/a99434b1117d651e5e40ebe7413d32f2db547788))
* **deps:** bump golang.org/x/text from 0.14.0 to 0.15.0 ([#466](https://github.com/rudderlabs/rudder-go-kit/issues/466)) ([816ce79](https://github.com/rudderlabs/rudder-go-kit/commit/816ce796b9e201f6ea73482a29f7b67d781e5183))

## [0.31.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.30.0...v0.31.0) (2024-05-17)


### Features

* add support for custom redis docker repository ([#467](https://github.com/rudderlabs/rudder-go-kit/issues/467)) ([55606f8](https://github.com/rudderlabs/rudder-go-kit/commit/55606f86ac4de5df387ae23407d1e347271a0612))


### Miscellaneous

* **deps:** bump github.com/prometheus/common from 0.52.3 to 0.53.0 ([#434](https://github.com/rudderlabs/rudder-go-kit/issues/434)) ([359fb64](https://github.com/rudderlabs/rudder-go-kit/commit/359fb6456ca22934e17fcbddd3532fa9ffa483b3))
* **deps:** bump go.etcd.io/etcd/client/v3 from 3.5.10 to 3.5.13 ([#445](https://github.com/rudderlabs/rudder-go-kit/issues/445)) ([af06f04](https://github.com/rudderlabs/rudder-go-kit/commit/af06f0456c83aa7b0a45c6ffd94d3535e05e94e6))
* **deps:** bump google.golang.org/api from 0.172.0 to 0.177.0 ([#459](https://github.com/rudderlabs/rudder-go-kit/issues/459)) ([352e560](https://github.com/rudderlabs/rudder-go-kit/commit/352e5608523c62ed2c45a5cc0e5cad5fcb154352))
* **deps:** bump the opentelemetry group with 9 updates ([#452](https://github.com/rudderlabs/rudder-go-kit/issues/452)) ([769e281](https://github.com/rudderlabs/rudder-go-kit/commit/769e28148203619dbf8be0f6414364121538acdd))

## [0.30.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.29.0...v0.30.0) (2024-04-30)


### Features

* add remoteFilePath support and create remote directory in sftp ([#437](https://github.com/rudderlabs/rudder-go-kit/issues/437)) ([cd85cf5](https://github.com/rudderlabs/rudder-go-kit/commit/cd85cf5a402cd1685b8aae5ddb2bb566a771aba0))


### Bug Fixes

* file creation on aws sftp servers ([#455](https://github.com/rudderlabs/rudder-go-kit/issues/455)) ([76d3b53](https://github.com/rudderlabs/rudder-go-kit/commit/76d3b53b019b32715e31280c8b652c869d53f5b1))


### Miscellaneous

* **deps:** bump google.golang.org/grpc from 1.63.0 to 1.63.2 ([#446](https://github.com/rudderlabs/rudder-go-kit/issues/446)) ([618baa6](https://github.com/rudderlabs/rudder-go-kit/commit/618baa64fa6c7d16ce3bd42d6d8c8c7e010b11b9))

## [0.29.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.28.0...v0.29.0) (2024-04-24)


### Features

* monitor sql database ([#444](https://github.com/rudderlabs/rudder-go-kit/issues/444)) ([ae1d227](https://github.com/rudderlabs/rudder-go-kit/commit/ae1d2274e740d1f4ee175d5e734c44b2784db611))

## [0.28.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.27.0...v0.28.0) (2024-04-23)


### Features

* add etcd resource [PIPE-971] ([#440](https://github.com/rudderlabs/rudder-go-kit/issues/440)) ([860b454](https://github.com/rudderlabs/rudder-go-kit/commit/860b4542352f4611ce77432e47a1d6017b0930b9))

## [0.27.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.26.0...v0.27.0) (2024-04-12)


### Features

* sync.Group ([#425](https://github.com/rudderlabs/rudder-go-kit/issues/425)) ([ac91461](https://github.com/rudderlabs/rudder-go-kit/commit/ac9146195f76492cf91aa32e34ab38ed01d2575d))


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.51.19 to 1.51.20 ([#428](https://github.com/rudderlabs/rudder-go-kit/issues/428)) ([7c58164](https://github.com/rudderlabs/rudder-go-kit/commit/7c581645f49cbf1179b191276cb60ce59d6d3ce2))

## [0.26.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.25.0...v0.26.0) (2024-04-12)


### Features

* throttling with retryAfter ([#422](https://github.com/rudderlabs/rudder-go-kit/issues/422)) ([c2904bf](https://github.com/rudderlabs/rudder-go-kit/commit/c2904bf38a1a635c60dc2a492db5080dc148aa33))


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.51.16 to 1.51.17 ([#420](https://github.com/rudderlabs/rudder-go-kit/issues/420)) ([c992d13](https://github.com/rudderlabs/rudder-go-kit/commit/c992d13dae16671b63a4974bcb8e2e70508c668d))
* **deps:** bump github.com/aws/aws-sdk-go from 1.51.17 to 1.51.19 ([#426](https://github.com/rudderlabs/rudder-go-kit/issues/426)) ([ce6f2ff](https://github.com/rudderlabs/rudder-go-kit/commit/ce6f2ffe2bb0849ec2dddde10e421e01f9c296f2))
* **deps:** bump github.com/prometheus/common from 0.52.2 to 0.52.3 ([#427](https://github.com/rudderlabs/rudder-go-kit/issues/427)) ([c9a5734](https://github.com/rudderlabs/rudder-go-kit/commit/c9a57344b0b4cec1cd1ca7ce5fcc7241a48212ea))

## [0.25.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.24.0...v0.25.0) (2024-04-09)


### Features

* additional helpers packages ([#416](https://github.com/rudderlabs/rudder-go-kit/issues/416)) ([877fd55](https://github.com/rudderlabs/rudder-go-kit/commit/877fd55912376e3422048584719cd332dc2c280f))


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.51.15 to 1.51.16 ([#418](https://github.com/rudderlabs/rudder-go-kit/issues/418)) ([bfdaa00](https://github.com/rudderlabs/rudder-go-kit/commit/bfdaa0015c4f565e7eccb69b00d191124910aec7))
* **deps:** bump golang.org/x/crypto from 0.21.0 to 0.22.0 ([#414](https://github.com/rudderlabs/rudder-go-kit/issues/414)) ([abf7a24](https://github.com/rudderlabs/rudder-go-kit/commit/abf7a2410e882091facab0153c00f43f33f366e8))
* **deps:** bump golang.org/x/oauth2 from 0.18.0 to 0.19.0 ([#413](https://github.com/rudderlabs/rudder-go-kit/issues/413)) ([08eef2f](https://github.com/rudderlabs/rudder-go-kit/commit/08eef2f78b8f86035db96f69819df7370c4884ce))
* **deps:** bump golang.org/x/sync from 0.6.0 to 0.7.0 ([#415](https://github.com/rudderlabs/rudder-go-kit/issues/415)) ([a19d647](https://github.com/rudderlabs/rudder-go-kit/commit/a19d64704d11ee723af3071751c64fb1afae784e))
* **deps:** bump the opentelemetry group with 9 updates ([#417](https://github.com/rudderlabs/rudder-go-kit/issues/417)) ([7344998](https://github.com/rudderlabs/rudder-go-kit/commit/73449986643f1801dfc1dd326ed02f8c418de691))
* provide methods to access config init errors ([#406](https://github.com/rudderlabs/rudder-go-kit/issues/406)) ([8579877](https://github.com/rudderlabs/rudder-go-kit/commit/85798776bfab15659066a6746bb2db9b020abe85))

## [0.24.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.23.3...v0.24.0) (2024-04-05)


### Features

* add sftp library ([#393](https://github.com/rudderlabs/rudder-go-kit/issues/393)) ([f0b67e9](https://github.com/rudderlabs/rudder-go-kit/commit/f0b67e93831d16b7f6618632ad44d718c8318b87))


### Bug Fixes

* health check for ssh server ([#411](https://github.com/rudderlabs/rudder-go-kit/issues/411)) ([c788d93](https://github.com/rudderlabs/rudder-go-kit/commit/c788d938775ccc9120244bf25d3726578b53c63b))


### Miscellaneous

* **deps:** bump cloud.google.com/go/storage from 1.39.1 to 1.40.0 ([#402](https://github.com/rudderlabs/rudder-go-kit/issues/402)) ([8ba1caf](https://github.com/rudderlabs/rudder-go-kit/commit/8ba1caf563d9189009951228209ddfa5ad21fabf))
* **deps:** bump github.com/aws/aws-sdk-go from 1.51.12 to 1.51.15 ([#412](https://github.com/rudderlabs/rudder-go-kit/issues/412)) ([6edd06e](https://github.com/rudderlabs/rudder-go-kit/commit/6edd06e9e3731463ce985a608df3eb6784f892ed))
* **deps:** bump github.com/aws/aws-sdk-go from 1.51.6 to 1.51.12 ([#405](https://github.com/rudderlabs/rudder-go-kit/issues/405)) ([47b4501](https://github.com/rudderlabs/rudder-go-kit/commit/47b4501852576bca15e0cefaefee415dbf208ed9))
* **deps:** bump github.com/cenkalti/backoff/v4 from 4.2.1 to 4.3.0 ([#394](https://github.com/rudderlabs/rudder-go-kit/issues/394)) ([17d8f10](https://github.com/rudderlabs/rudder-go-kit/commit/17d8f10783927f719270a39f5e844d352e843048))
* **deps:** bump github.com/prometheus/client_model from 0.6.0 to 0.6.1 ([#408](https://github.com/rudderlabs/rudder-go-kit/issues/408)) ([1f6d86d](https://github.com/rudderlabs/rudder-go-kit/commit/1f6d86d3b76cae98c0cf825c4000314acd95ec22))
* **deps:** bump github.com/prometheus/common from 0.51.1 to 0.52.2 ([#409](https://github.com/rudderlabs/rudder-go-kit/issues/409)) ([d930ce3](https://github.com/rudderlabs/rudder-go-kit/commit/d930ce3185225d095ae3e8f70fec18cd77a75182))
* **deps:** bump github.com/shirou/gopsutil/v3 from 3.24.2 to 3.24.3 ([#404](https://github.com/rudderlabs/rudder-go-kit/issues/404)) ([6c33058](https://github.com/rudderlabs/rudder-go-kit/commit/6c33058f7f842f7a1fd345fe4f0a5924936e188b))
* **deps:** bump google.golang.org/api from 0.171.0 to 0.172.0 ([#398](https://github.com/rudderlabs/rudder-go-kit/issues/398)) ([085ffc7](https://github.com/rudderlabs/rudder-go-kit/commit/085ffc75b92ddea8ae0b8ab596f1045551d39137))

## [0.23.3](https://github.com/rudderlabs/rudder-go-kit/compare/v0.23.2...v0.23.3) (2024-03-25)


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.50.38 to 1.51.6 ([#391](https://github.com/rudderlabs/rudder-go-kit/issues/391)) ([be9d6c1](https://github.com/rudderlabs/rudder-go-kit/commit/be9d6c1ac86bf38c9f0614b102767cfd2e92ae43))
* **deps:** bump github.com/docker/docker from 20.10.27+incompatible to 24.0.9+incompatible ([#385](https://github.com/rudderlabs/rudder-go-kit/issues/385)) ([1ce09ad](https://github.com/rudderlabs/rudder-go-kit/commit/1ce09adb73211eaad26a2ed79cc269611a7234b9))
* **deps:** bump github.com/prometheus/common from 0.50.0 to 0.51.1 ([#392](https://github.com/rudderlabs/rudder-go-kit/issues/392)) ([3f4ce58](https://github.com/rudderlabs/rudder-go-kit/commit/3f4ce58c076d89b22964ff48ad5982cc0b221f87))
* **deps:** bump google.golang.org/api from 0.167.0 to 0.171.0 ([#389](https://github.com/rudderlabs/rudder-go-kit/issues/389)) ([4f23940](https://github.com/rudderlabs/rudder-go-kit/commit/4f23940b0dc518f5c0120fd2dd769de13592dff9))

## [0.23.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.23.1...v0.23.2) (2024-03-15)


### Bug Fixes

* indefinite function runs inside mutex lock ([#372](https://github.com/rudderlabs/rudder-go-kit/issues/372)) ([0c2c23b](https://github.com/rudderlabs/rudder-go-kit/commit/0c2c23b8fdc319163d6a3570de6e89ff7833ef53))


### Miscellaneous

* **deps:** bump cloud.google.com/go/storage from 1.38.0 to 1.39.1 ([#377](https://github.com/rudderlabs/rudder-go-kit/issues/377)) ([a520f91](https://github.com/rudderlabs/rudder-go-kit/commit/a520f915165cc5b6ce2e8081d2cf5b030c7e88fe))
* **deps:** bump github.com/aws/aws-sdk-go from 1.50.24 to 1.50.30 ([#366](https://github.com/rudderlabs/rudder-go-kit/issues/366)) ([6255c36](https://github.com/rudderlabs/rudder-go-kit/commit/6255c36051d664b1d58b19462913735261809369))
* **deps:** bump github.com/aws/aws-sdk-go from 1.50.30 to 1.50.38 ([#378](https://github.com/rudderlabs/rudder-go-kit/issues/378)) ([dfefeec](https://github.com/rudderlabs/rudder-go-kit/commit/dfefeecc9ad7c58fdb22147fdc458f99584b9f10))
* **deps:** bump github.com/minio/minio-go/v7 from 7.0.67 to 7.0.69 ([#373](https://github.com/rudderlabs/rudder-go-kit/issues/373)) ([77ee989](https://github.com/rudderlabs/rudder-go-kit/commit/77ee9890647c1c04a90af60514a8441f53c625e5))
* **deps:** bump github.com/prometheus/common from 0.48.0 to 0.50.0 ([#371](https://github.com/rudderlabs/rudder-go-kit/issues/371)) ([4c6d4b6](https://github.com/rudderlabs/rudder-go-kit/commit/4c6d4b6169a15dbe930e2d87958a2962fc1a99b4))
* **deps:** bump github.com/shirou/gopsutil/v3 from 3.24.1 to 3.24.2 ([#369](https://github.com/rudderlabs/rudder-go-kit/issues/369)) ([1b4cc97](https://github.com/rudderlabs/rudder-go-kit/commit/1b4cc9725fabc6ddddabb9d71f0d102de1553915))
* **deps:** bump github.com/stretchr/testify from 1.8.4 to 1.9.0 ([#379](https://github.com/rudderlabs/rudder-go-kit/issues/379)) ([0ec92af](https://github.com/rudderlabs/rudder-go-kit/commit/0ec92afe6a6cb7a19af3271190fd58849d9f3024))
* **deps:** bump google.golang.org/api from 0.166.0 to 0.167.0 ([#356](https://github.com/rudderlabs/rudder-go-kit/issues/356)) ([6982111](https://github.com/rudderlabs/rudder-go-kit/commit/69821111eea4b85afbe29c5891b517cde3acf466))
* **deps:** bump the opentelemetry group with 9 updates ([#354](https://github.com/rudderlabs/rudder-go-kit/issues/354)) ([f4fb7b4](https://github.com/rudderlabs/rudder-go-kit/commit/f4fb7b4c1f36503e361952d7b305dcc840d4249c))
* export config BE handler for transformer, and unstarted http test server ([#374](https://github.com/rudderlabs/rudder-go-kit/issues/374)) ([018b6f7](https://github.com/rudderlabs/rudder-go-kit/commit/018b6f742d4b65be1f9fd3b453e083a79470999f))

## [0.23.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.23.0...v0.23.1) (2024-02-27)


### Miscellaneous

* close postgres db on cleanup ([#357](https://github.com/rudderlabs/rudder-go-kit/issues/357)) ([85f29b5](https://github.com/rudderlabs/rudder-go-kit/commit/85f29b52ecdece8ecca3002402e1c6cf8f154001))

## [0.23.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.22.0...v0.23.0) (2024-02-23)


### Features

* add async.SingleSender and a new docker resource for mysql ([#333](https://github.com/rudderlabs/rudder-go-kit/issues/333)) ([063ec10](https://github.com/rudderlabs/rudder-go-kit/commit/063ec102da7e1431dd0c27e45c46239a6ac67d40))


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.50.23 to 1.50.24 ([#352](https://github.com/rudderlabs/rudder-go-kit/issues/352)) ([1515334](https://github.com/rudderlabs/rudder-go-kit/commit/15153344b7bb9d9e6291c8309711705a475600ea))
* **deps:** bump github.com/prometheus/common from 0.47.0 to 0.48.0 ([#351](https://github.com/rudderlabs/rudder-go-kit/issues/351)) ([f44e67c](https://github.com/rudderlabs/rudder-go-kit/commit/f44e67c4c9fc633fd177c5c5d7918c6d5766fb7c))

## [0.22.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.21.1...v0.22.0) (2024-02-23)


### Features

* add mock config BE support for transformer ([#344](https://github.com/rudderlabs/rudder-go-kit/issues/344)) ([d48fc68](https://github.com/rudderlabs/rudder-go-kit/commit/d48fc68b3dc78f1be50c73aa3353a8dd4ab76821))


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.50.21 to 1.50.23 ([#348](https://github.com/rudderlabs/rudder-go-kit/issues/348)) ([c4ec5d6](https://github.com/rudderlabs/rudder-go-kit/commit/c4ec5d62a9401df4bcc953cbf9513b24a747a2ab))
* **deps:** bump go.uber.org/zap from 1.26.0 to 1.27.0 ([#346](https://github.com/rudderlabs/rudder-go-kit/issues/346)) ([fba3feb](https://github.com/rudderlabs/rudder-go-kit/commit/fba3feb3cd7b7e0697223ef3e22a871ae1afbb14))
* **deps:** bump google.golang.org/api from 0.165.0 to 0.166.0 ([#349](https://github.com/rudderlabs/rudder-go-kit/issues/349)) ([659369d](https://github.com/rudderlabs/rudder-go-kit/commit/659369d50b731c05f305f839510925144f5418e2))

## [0.21.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.21.0...v0.21.1) (2024-02-21)


### Miscellaneous

* **deps:** bump cloud.google.com/go/storage from 1.37.0 to 1.38.0 ([#331](https://github.com/rudderlabs/rudder-go-kit/issues/331)) ([e1b23a4](https://github.com/rudderlabs/rudder-go-kit/commit/e1b23a4897bed970b647ddb08614d11f9c2936c5))
* **deps:** bump github.com/aws/aws-sdk-go from 1.50.13 to 1.50.17 ([#335](https://github.com/rudderlabs/rudder-go-kit/issues/335)) ([ffbf13c](https://github.com/rudderlabs/rudder-go-kit/commit/ffbf13c4b7ce5f5ef55450e03805bbb63913d254))
* **deps:** bump github.com/aws/aws-sdk-go from 1.50.17 to 1.50.19 ([#338](https://github.com/rudderlabs/rudder-go-kit/issues/338)) ([ebdee1f](https://github.com/rudderlabs/rudder-go-kit/commit/ebdee1f1858c37fbee7b87fdc97be8cf1d4849e8))
* **deps:** bump github.com/aws/aws-sdk-go from 1.50.19 to 1.50.21 ([#343](https://github.com/rudderlabs/rudder-go-kit/issues/343)) ([60169d5](https://github.com/rudderlabs/rudder-go-kit/commit/60169d56ecb3acb4e11da4a634f9282cbaae0e36))
* **deps:** bump github.com/go-chi/chi/v5 from 5.0.11 to 5.0.12 ([#340](https://github.com/rudderlabs/rudder-go-kit/issues/340)) ([793799f](https://github.com/rudderlabs/rudder-go-kit/commit/793799f356b5511b2dc55a4e50d4ae4fec6e3e5f))
* **deps:** bump github.com/minio/minio-go/v7 from 7.0.66 to 7.0.67 ([#329](https://github.com/rudderlabs/rudder-go-kit/issues/329)) ([d65d0b8](https://github.com/rudderlabs/rudder-go-kit/commit/d65d0b86d273a1667d5320c3d97a024c7955cfe2))
* **deps:** bump github.com/prometheus/client_model from 0.5.0 to 0.6.0 ([#342](https://github.com/rudderlabs/rudder-go-kit/issues/342)) ([fab0397](https://github.com/rudderlabs/rudder-go-kit/commit/fab0397649fd4af65974bdd3fbd2c1ee84dff281))
* **deps:** bump github.com/prometheus/common from 0.46.0 to 0.47.0 ([#339](https://github.com/rudderlabs/rudder-go-kit/issues/339)) ([55efbad](https://github.com/rudderlabs/rudder-go-kit/commit/55efbadf66b99095175cbcbfbb3faccdefbb7cfe))
* **deps:** bump golang.org/x/oauth2 from 0.16.0 to 0.17.0 ([#323](https://github.com/rudderlabs/rudder-go-kit/issues/323)) ([331cc2f](https://github.com/rudderlabs/rudder-go-kit/commit/331cc2f6a40332f04fa2d81100560bca27ced9ca))
* **deps:** bump google.golang.org/api from 0.162.0 to 0.165.0 ([#337](https://github.com/rudderlabs/rudder-go-kit/issues/337)) ([9f118b2](https://github.com/rudderlabs/rudder-go-kit/commit/9f118b2ce18ce4309df93b2e5c716da79050eb1c))
* race support ([#345](https://github.com/rudderlabs/rudder-go-kit/issues/345)) ([f8cb291](https://github.com/rudderlabs/rudder-go-kit/commit/f8cb291075af5ed37bfc65d162a1b80a518f76ed))

## [0.21.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.20.3...v0.21.0) (2024-02-15)


### Features

* add ack to kafka client ([#327](https://github.com/rudderlabs/rudder-go-kit/issues/327)) ([b4e3c34](https://github.com/rudderlabs/rudder-go-kit/commit/b4e3c34b07ae1b3ba73c7dbd4af379c0b20f7f95))


### Miscellaneous

* add docker resource for transformer ([#334](https://github.com/rudderlabs/rudder-go-kit/issues/334)) ([3c98aaf](https://github.com/rudderlabs/rudder-go-kit/commit/3c98aafd6f08ba5e00200cb86dbe0a02ae4cdfec))
* **deps:** bump the opentelemetry group with 9 updates ([#322](https://github.com/rudderlabs/rudder-go-kit/issues/322)) ([111bd3b](https://github.com/rudderlabs/rudder-go-kit/commit/111bd3b3038448c835d091c077de27de6e38dc99))

## [0.20.3](https://github.com/rudderlabs/rudder-go-kit/compare/v0.20.2...v0.20.3) (2024-02-08)


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.50.6 to 1.50.13 ([#321](https://github.com/rudderlabs/rudder-go-kit/issues/321)) ([596e29b](https://github.com/rudderlabs/rudder-go-kit/commit/596e29b826e8eb120830af6eeeda37a7c861fd55))
* **deps:** bump github.com/opencontainers/runc from 1.1.5 to 1.1.12 ([#309](https://github.com/rudderlabs/rudder-go-kit/issues/309)) ([669e348](https://github.com/rudderlabs/rudder-go-kit/commit/669e348887e7ef1f63500afe438f9444187aeac1))
* **deps:** bump github.com/shirou/gopsutil/v3 from 3.23.12 to 3.24.1 ([#312](https://github.com/rudderlabs/rudder-go-kit/issues/312)) ([7feeef4](https://github.com/rudderlabs/rudder-go-kit/commit/7feeef4535b383bb1c3d190d2cf229421376cdcd))
* **deps:** bump golang.org/x/crypto from 0.18.0 to 0.19.0 ([#320](https://github.com/rudderlabs/rudder-go-kit/issues/320)) ([3f0d228](https://github.com/rudderlabs/rudder-go-kit/commit/3f0d2283d1ee33a7e9930a2f48045c8ca4b2b039))
* **deps:** bump google.golang.org/api from 0.160.0 to 0.162.0 ([#315](https://github.com/rudderlabs/rudder-go-kit/issues/315)) ([84fa993](https://github.com/rudderlabs/rudder-go-kit/commit/84fa993a5ef5bcc8b9a9c510a7ab2c7ccb5787b5))
* unexpected EOF errors with postgres container ([#319](https://github.com/rudderlabs/rudder-go-kit/issues/319)) ([0d8857b](https://github.com/rudderlabs/rudder-go-kit/commit/0d8857ba6fc3e8181588c00a518bfdc51d1bf86e))

## [0.20.2](https://github.com/rudderlabs/rudder-go-kit/compare/v0.20.1...v0.20.2) (2024-01-30)


### Miscellaneous

* **deps:** bump github.com/docker/docker from 20.10.17+incompatible to 20.10.27+incompatible ([#305](https://github.com/rudderlabs/rudder-go-kit/issues/305)) ([460b812](https://github.com/rudderlabs/rudder-go-kit/commit/460b812cce81302205992c233e2df9f8a3ca2237))

## [0.20.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.20.0...v0.20.1) (2024-01-30)


### Miscellaneous

* go mod tidy ([#303](https://github.com/rudderlabs/rudder-go-kit/issues/303)) ([9997004](https://github.com/rudderlabs/rudder-go-kit/commit/999700448e7d24dd589fce94ac9b7aa5456b6e02))

## [0.20.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.19.1...v0.20.0) (2024-01-30)


### Features

* kafka client and containers ([#287](https://github.com/rudderlabs/rudder-go-kit/issues/287)) ([00cba9d](https://github.com/rudderlabs/rudder-go-kit/commit/00cba9dbb04b6ff9807dcf6a3d2a35acf8137b49))


### Miscellaneous

* adding sdk to otel group ([#301](https://github.com/rudderlabs/rudder-go-kit/issues/301)) ([7385e0d](https://github.com/rudderlabs/rudder-go-kit/commit/7385e0db3c14ab99e922d97db3ddc392c48b8997))
* dependabot otel ([#296](https://github.com/rudderlabs/rudder-go-kit/issues/296)) ([67b2cc4](https://github.com/rudderlabs/rudder-go-kit/commit/67b2cc40fb946214087f4f229fa44662786afaa1))
* **deps:** bump github.com/aws/aws-sdk-go from 1.49.21 to 1.49.24 ([#293](https://github.com/rudderlabs/rudder-go-kit/issues/293)) ([f220d47](https://github.com/rudderlabs/rudder-go-kit/commit/f220d47f85adeb9993de47cc139e37f9bea505f8))
* **deps:** bump the opentelemetry group with 5 updates ([#297](https://github.com/rudderlabs/rudder-go-kit/issues/297)) ([781ad78](https://github.com/rudderlabs/rudder-go-kit/commit/781ad78d96047b39314d69094bcdad3320816f95))

## [0.19.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.19.0...v0.19.1) (2024-01-16)


### Miscellaneous

* **deps:** bump cloud.google.com/go/storage from 1.35.1 to 1.36.0 ([#256](https://github.com/rudderlabs/rudder-go-kit/issues/256)) ([4b52787](https://github.com/rudderlabs/rudder-go-kit/commit/4b527871b246a5d5f64cb27b65c0969d83900df7))
* **deps:** bump github.com/aws/aws-sdk-go from 1.48.11 to 1.49.3 ([#254](https://github.com/rudderlabs/rudder-go-kit/issues/254)) ([02553b0](https://github.com/rudderlabs/rudder-go-kit/commit/02553b005fa808003eaf0f508ecd40c5a37c7c58))
* **deps:** bump github.com/aws/aws-sdk-go from 1.49.13 to 1.49.21 ([#284](https://github.com/rudderlabs/rudder-go-kit/issues/284)) ([2374e4e](https://github.com/rudderlabs/rudder-go-kit/commit/2374e4e5b253e4b248206df3be36eda119d4b9ce))
* **deps:** bump github.com/aws/aws-sdk-go from 1.49.3 to 1.49.13 ([#272](https://github.com/rudderlabs/rudder-go-kit/issues/272)) ([29b58db](https://github.com/rudderlabs/rudder-go-kit/commit/29b58db5479ff928a153e139429dfd702f1ae72c))
* **deps:** bump github.com/go-chi/chi/v5 from 5.0.10 to 5.0.11 ([#264](https://github.com/rudderlabs/rudder-go-kit/issues/264)) ([9352abb](https://github.com/rudderlabs/rudder-go-kit/commit/9352abbefe6f701b18a04144d0dc73ba98bee891))
* **deps:** bump github.com/minio/minio-go/v7 from 7.0.63 to 7.0.65 ([#239](https://github.com/rudderlabs/rudder-go-kit/issues/239)) ([35377d2](https://github.com/rudderlabs/rudder-go-kit/commit/35377d2edc88f9820809fcc19d86e666538d52e5))
* **deps:** bump github.com/minio/minio-go/v7 from 7.0.65 to 7.0.66 ([#257](https://github.com/rudderlabs/rudder-go-kit/issues/257)) ([06ded85](https://github.com/rudderlabs/rudder-go-kit/commit/06ded8555711681833dafba4b74400a282ee7082))
* **deps:** bump github.com/prometheus/client_golang from 1.17.0 to 1.18.0 ([#268](https://github.com/rudderlabs/rudder-go-kit/issues/268)) ([0131533](https://github.com/rudderlabs/rudder-go-kit/commit/01315335f9508f2d8eb8565b9316a30a6a0553aa))
* **deps:** bump github.com/prometheus/common from 0.45.0 to 0.46.0 ([#285](https://github.com/rudderlabs/rudder-go-kit/issues/285)) ([dffa4a3](https://github.com/rudderlabs/rudder-go-kit/commit/dffa4a3a76897189403e7ac3e74cd5d78f8136ea))
* **deps:** bump github.com/samber/lo from 1.38.1 to 1.39.0 ([#246](https://github.com/rudderlabs/rudder-go-kit/issues/246)) ([2ecde55](https://github.com/rudderlabs/rudder-go-kit/commit/2ecde5526b0e5f0e637ec0bf35708bfa714217dd))
* **deps:** bump github.com/shirou/gopsutil/v3 from 3.23.10 to 3.23.11 ([#247](https://github.com/rudderlabs/rudder-go-kit/issues/247)) ([5b86382](https://github.com/rudderlabs/rudder-go-kit/commit/5b8638271d8db8a479273c28b2214b270467e7a5))
* **deps:** bump github.com/shirou/gopsutil/v3 from 3.23.11 to 3.23.12 ([#271](https://github.com/rudderlabs/rudder-go-kit/issues/271)) ([264aeff](https://github.com/rudderlabs/rudder-go-kit/commit/264aeff695c5ab7a913817c5b0b2636a8cf20eef))
* **deps:** bump github.com/spf13/cast from 1.5.1 to 1.6.0 ([#233](https://github.com/rudderlabs/rudder-go-kit/issues/233)) ([230fb2a](https://github.com/rudderlabs/rudder-go-kit/commit/230fb2ad6594a0b7dca350d6a2fd736d0e65a067))
* **deps:** bump github.com/spf13/viper from 1.17.0 to 1.18.1 ([#248](https://github.com/rudderlabs/rudder-go-kit/issues/248)) ([558fe3c](https://github.com/rudderlabs/rudder-go-kit/commit/558fe3c6311be3b7b1c4e1178740974ca5a8f1e6))
* **deps:** bump github.com/spf13/viper from 1.18.1 to 1.18.2 ([#269](https://github.com/rudderlabs/rudder-go-kit/issues/269)) ([afd2a42](https://github.com/rudderlabs/rudder-go-kit/commit/afd2a428a33f60cb9cc51eb21505c81450037942))
* **deps:** bump golang.org/x/crypto from 0.16.0 to 0.17.0 ([#258](https://github.com/rudderlabs/rudder-go-kit/issues/258)) ([556a590](https://github.com/rudderlabs/rudder-go-kit/commit/556a59089877492a76862c1b0ebe5b88ff0048ea))
* **deps:** bump golang.org/x/oauth2 from 0.14.0 to 0.15.0 ([#241](https://github.com/rudderlabs/rudder-go-kit/issues/241)) ([be46957](https://github.com/rudderlabs/rudder-go-kit/commit/be469578280a82f6340d2c696ceb47f5856e5cae))
* **deps:** bump golang.org/x/oauth2 from 0.15.0 to 0.16.0 ([#280](https://github.com/rudderlabs/rudder-go-kit/issues/280)) ([e59d900](https://github.com/rudderlabs/rudder-go-kit/commit/e59d9001400e1e830f3a9d2be6477e750e0a9bab))
* **deps:** bump golang.org/x/sync from 0.5.0 to 0.6.0 ([#276](https://github.com/rudderlabs/rudder-go-kit/issues/276)) ([9680814](https://github.com/rudderlabs/rudder-go-kit/commit/968081407280b77de8f89fe81fdaf9c0948b8ec4))
* **deps:** bump google.golang.org/api from 0.153.0 to 0.154.0 ([#251](https://github.com/rudderlabs/rudder-go-kit/issues/251)) ([bb154bc](https://github.com/rudderlabs/rudder-go-kit/commit/bb154bcb6dc96112df978d0950ec543b006d202d))
* **deps:** bump google.golang.org/api from 0.154.0 to 0.156.0 ([#282](https://github.com/rudderlabs/rudder-go-kit/issues/282)) ([425c1b8](https://github.com/rudderlabs/rudder-go-kit/commit/425c1b8bafc550b12b9cb1af708242ace20fbb38))
* **deps:** bump google.golang.org/protobuf from 1.31.0 to 1.32.0 ([#267](https://github.com/rudderlabs/rudder-go-kit/issues/267)) ([32287c9](https://github.com/rudderlabs/rudder-go-kit/commit/32287c94196b88876b5079e670ff668cc11fe6e0))

## [0.19.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.18.1...v0.19.0) (2023-12-07)


### Features

* tracing support ([#231](https://github.com/rudderlabs/rudder-go-kit/issues/231)) ([e9a89cb](https://github.com/rudderlabs/rudder-go-kit/commit/e9a89cb152569af509f1a8c17e6b5ced1abce809))


### Miscellaneous

* dependabot otel group ([#221](https://github.com/rudderlabs/rudder-go-kit/issues/221)) ([fb71ca8](https://github.com/rudderlabs/rudder-go-kit/commit/fb71ca81f8ea4d6a5962dfbfb4ffed361f7c0c2e))
* **deps:** bump cloud.google.com/go/storage from 1.34.1 to 1.35.1 ([#222](https://github.com/rudderlabs/rudder-go-kit/issues/222)) ([678bc3b](https://github.com/rudderlabs/rudder-go-kit/commit/678bc3b3969c5ef03d1c539c42e7f750514fd37d))
* **deps:** bump github.com/aws/aws-sdk-go from 1.47.10 to 1.48.11 ([#240](https://github.com/rudderlabs/rudder-go-kit/issues/240)) ([abf3583](https://github.com/rudderlabs/rudder-go-kit/commit/abf3583f84a1c21f5c6f63ade72ce4e411586a72))
* **deps:** bump google.golang.org/api from 0.150.0 to 0.153.0 ([#242](https://github.com/rudderlabs/rudder-go-kit/issues/242)) ([628c035](https://github.com/rudderlabs/rudder-go-kit/commit/628c0357b83a5fa9f946aacc4dcd8b9b96d0000c))
* golangci-lint no docker ([#228](https://github.com/rudderlabs/rudder-go-kit/issues/228)) ([080bc69](https://github.com/rudderlabs/rudder-go-kit/commit/080bc69aa5dde0131c9d0b8dd6f49f12f016e0ee))

## [0.18.1](https://github.com/rudderlabs/rudder-go-kit/compare/v0.18.0...v0.18.1) (2023-11-21)


### Miscellaneous

* update otel ([#219](https://github.com/rudderlabs/rudder-go-kit/issues/219)) ([7a59061](https://github.com/rudderlabs/rudder-go-kit/commit/7a59061743a17da77a580a96194bfab569b221c0))

## [0.18.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.17.0...v0.18.0) (2023-11-17)


### Features

* fields for labels and non-sugared structured logging ([#202](https://github.com/rudderlabs/rudder-go-kit/issues/202)) ([205700e](https://github.com/rudderlabs/rudder-go-kit/commit/205700ee149520d093a4e2af556f00e097776e77))

## [0.17.0](https://github.com/rudderlabs/rudder-go-kit/compare/v0.16.4...v0.17.0) (2023-11-16)


### Features

* provide memstat methods to assist on testing (pipe-538) ([#210](https://github.com/rudderlabs/rudder-go-kit/issues/210)) ([8150925](https://github.com/rudderlabs/rudder-go-kit/commit/81509250ac6dc58b3d778559e38491a43ce52ae0))


### Miscellaneous

* **deps:** bump github.com/aws/aws-sdk-go from 1.47.8 to 1.47.10 ([#208](https://github.com/rudderlabs/rudder-go-kit/issues/208)) ([52204ef](https://github.com/rudderlabs/rudder-go-kit/commit/52204efc12adf4d302b2857fb7c995f73cfb8cd3))

## [0.16.4](https://github.com/rudderlabs/rudder-go-kit/compare/v0.16.3...v0.16.4) (2023-11-10)


### Miscellaneous

* **deps:** bump cloud.google.com/go/storage from 1.33.0 to 1.34.1 ([#191](https://github.com/rudderlabs/rudder-go-kit/issues/191)) ([18577aa](https://github.com/rudderlabs/rudder-go-kit/commit/18577aa5ed21f78d491d0ad487fd1cf260be1ee4))
* **deps:** bump github.com/aws/aws-sdk-go from 1.46.4 to 1.47.3 ([#193](https://github.com/rudderlabs/rudder-go-kit/issues/193)) ([f25a511](https://github.com/rudderlabs/rudder-go-kit/commit/f25a511d2dc4c809a56d1aae286c562eb687e069))
* **deps:** bump github.com/aws/aws-sdk-go from 1.47.3 to 1.47.8 ([#201](https://github.com/rudderlabs/rudder-go-kit/issues/201)) ([47e0183](https://github.com/rudderlabs/rudder-go-kit/commit/47e018364b9273830ef5d6d49d5cd136dd0c6b67))
* **deps:** bump github.com/docker/docker from 20.10.24+incompatible to 24.0.7+incompatible ([#184](https://github.com/rudderlabs/rudder-go-kit/issues/184)) ([abe3e0a](https://github.com/rudderlabs/rudder-go-kit/commit/abe3e0a75383a5d3a764442592238d64fccafa3f))
* **deps:** bump github.com/shirou/gopsutil/v3 from 3.23.9 to 3.23.10 ([#187](https://github.com/rudderlabs/rudder-go-kit/issues/187)) ([089f784](https://github.com/rudderlabs/rudder-go-kit/commit/089f784b7d4a89307e5f69aa94fcc1ea5974a091))
* **deps:** bump golang.org/x/oauth2 from 0.13.0 to 0.14.0 ([#200](https://github.com/rudderlabs/rudder-go-kit/issues/200)) ([db0d9c6](https://github.com/rudderlabs/rudder-go-kit/commit/db0d9c67b86de35194ef35c4c597cddca13d39cd))
* **deps:** bump google.golang.org/api from 0.148.0 to 0.149.0 ([#189](https://github.com/rudderlabs/rudder-go-kit/issues/189)) ([03e48e0](https://github.com/rudderlabs/rudder-go-kit/commit/03e48e0efc96c36265ded5bc084739ada489b841))
* **deps:** bump google.golang.org/api from 0.148.0 to 0.150.0 ([#195](https://github.com/rudderlabs/rudder-go-kit/issues/195)) ([a479231](https://github.com/rudderlabs/rudder-go-kit/commit/a479231e58191945981ea99df03b90823503a902))

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
