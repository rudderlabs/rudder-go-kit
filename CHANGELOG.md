# Changelog

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
