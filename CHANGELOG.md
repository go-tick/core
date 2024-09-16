# [0.1.0-beta.1](https://github.com/misikdmytro/gotick/compare/v0.0.1...v0.1.0-beta.1) (2024-09-16)


### Bug Fixes

* bypass SchedulerID to job context & improve tests for this ([bd947d6](https://github.com/misikdmytro/gotick/commit/bd947d65985376b5ed9da804d3d9b93ccd49fe49))
* fix driver registration problem with returning pointer instead of value itself ([361a66c](https://github.com/misikdmytro/gotick/commit/361a66cfe6ffa2463c1d3f5425a913016636e528))
* fix lint issues ([ca15127](https://github.com/misikdmytro/gotick/commit/ca151273dcb17335c633e9b2ca36268ba3fa8cfa))
* fix lint issues (2) ([18dfa80](https://github.com/misikdmytro/gotick/commit/18dfa803f591bddc2c904dc112b6036e605d6f60))
* fix nil pointer dereference ([97a10a1](https://github.com/misikdmytro/gotick/commit/97a10a14fe1b42c772392af23345368f389c0b0d))
* mark job as currently executed in driver after the event from scheduler ([690fc64](https://github.com/misikdmytro/gotick/commit/690fc6485c70b2862fa1e39125c87b01106b3f5e))
* race condition for in-memory driver ([81f9eaa](https://github.com/misikdmytro/gotick/commit/81f9eaaae9cffd8d9ef62fabf0dae590d2aec53a))


### Features

* allow multi-threaded scheduler ([c5025a1](https://github.com/misikdmytro/gotick/commit/c5025a1cc31cb8eb8f2015dd1d6f8e939a12ab26))
* custom jobs timeout ([8f80dd1](https://github.com/misikdmytro/gotick/commit/8f80dd15ad64b3f44a1ddd3deefe2fa1f109504d))
* implement in memory driver (without tests for now) [skip ci] ([13e01dc](https://github.com/misikdmytro/gotick/commit/13e01dcf6f5e54e3b4937563a5d643d689f73638))
* implement job planner [skip ci] ([d05bc17](https://github.com/misikdmytro/gotick/commit/d05bc17c6b2a3b2f68e88e438f06ccf56ac0d2ec))
* implement job scheduling policies ([2181887](https://github.com/misikdmytro/gotick/commit/21818876bb979556450bcae23a28c1492cbb43d4))
* refactor planner &scheduler + new tests ([243ad99](https://github.com/misikdmytro/gotick/commit/243ad99796a605346efb309e94f0228e694fdfa1))
* remove errors propagation from subscriber _. simplify code ([8ecc04a](https://github.com/misikdmytro/gotick/commit/8ecc04a8d8900ac04468e068d6c16d3386fbbc4a))
* simplify interface by removing propagating errors and extend subscriber activity ([84cd6a5](https://github.com/misikdmytro/gotick/commit/84cd6a575dc9a3adbfb5501878c562d138d6e7b5))
* start implementing scheduler [skip ci] ([64eb2fb](https://github.com/misikdmytro/gotick/commit/64eb2fb7990ad808855f748eb68eeb3122c3da58))
* test CD ([d1e566e](https://github.com/misikdmytro/gotick/commit/d1e566e87ed636983f5c382e38684cd5d282d6e0))
* update driver & factory registration ([0ada207](https://github.com/misikdmytro/gotick/commit/0ada207eec625aa58cffb0e5d235a8b46fd0a9b0))
* update exported modules from the package ([083224f](https://github.com/misikdmytro/gotick/commit/083224f06f5a81e2780fc60196bd02bdf9b8af4d))
* update husky to allow commit message length to be 1000 symbols ([6ffc680](https://github.com/misikdmytro/gotick/commit/6ffc680194efe51d07d6124eed53351ba79c6028))
* update planner & scheduler & cover planner with tests ([6a2e303](https://github.com/misikdmytro/gotick/commit/6a2e30335ec449fd2a02757db6c0c5e60cea8b5c))
* update public API, update settings & implement integration tests ([c9a679f](https://github.com/misikdmytro/gotick/commit/c9a679ff47a30a92bb8ae01bea1b5f72e8af709f))
* update schedule interface [skip ci] ([1159453](https://github.com/misikdmytro/gotick/commit/115945383855660926e67c5ec5db0cfab9e62346))
* update scheduler interface & cover with more tests ([853cbbf](https://github.com/misikdmytro/gotick/commit/853cbbfd529f7b1e2e1b2f54030d7e63eace5c5a))
* update schedules ([0c507f9](https://github.com/misikdmytro/gotick/commit/0c507f9135cd9d039d47b1a8e4c2cf9446284247))

# [0.1.0](https://github.com/misikdmytro/gotick/compare/v0.0.1...v0.1.0) (2024-09-15)


### Bug Fixes

* bypass SchedulerID to job context & improve tests for this ([bd947d6](https://github.com/misikdmytro/gotick/commit/bd947d65985376b5ed9da804d3d9b93ccd49fe49))
* fix driver registration problem with returning pointer instead of value itself ([361a66c](https://github.com/misikdmytro/gotick/commit/361a66cfe6ffa2463c1d3f5425a913016636e528))
* fix lint issues ([ca15127](https://github.com/misikdmytro/gotick/commit/ca151273dcb17335c633e9b2ca36268ba3fa8cfa))
* fix lint issues (2) ([18dfa80](https://github.com/misikdmytro/gotick/commit/18dfa803f591bddc2c904dc112b6036e605d6f60))
* fix nil pointer dereference ([97a10a1](https://github.com/misikdmytro/gotick/commit/97a10a14fe1b42c772392af23345368f389c0b0d))
* mark job as currently executed in driver after the event from scheduler ([690fc64](https://github.com/misikdmytro/gotick/commit/690fc6485c70b2862fa1e39125c87b01106b3f5e))
* race condition for in-memory driver ([81f9eaa](https://github.com/misikdmytro/gotick/commit/81f9eaaae9cffd8d9ef62fabf0dae590d2aec53a))


### Features

* allow multi-threaded scheduler ([c5025a1](https://github.com/misikdmytro/gotick/commit/c5025a1cc31cb8eb8f2015dd1d6f8e939a12ab26))
* custom jobs timeout ([8f80dd1](https://github.com/misikdmytro/gotick/commit/8f80dd15ad64b3f44a1ddd3deefe2fa1f109504d))
* implement in memory driver (without tests for now) [skip ci] ([13e01dc](https://github.com/misikdmytro/gotick/commit/13e01dcf6f5e54e3b4937563a5d643d689f73638))
* implement job planner [skip ci] ([d05bc17](https://github.com/misikdmytro/gotick/commit/d05bc17c6b2a3b2f68e88e438f06ccf56ac0d2ec))
* implement job scheduling policies ([2181887](https://github.com/misikdmytro/gotick/commit/21818876bb979556450bcae23a28c1492cbb43d4))
* refactor planner &scheduler + new tests ([243ad99](https://github.com/misikdmytro/gotick/commit/243ad99796a605346efb309e94f0228e694fdfa1))
* remove errors propagation from subscriber _. simplify code ([8ecc04a](https://github.com/misikdmytro/gotick/commit/8ecc04a8d8900ac04468e068d6c16d3386fbbc4a))
* simplify interface by removing propagating errors and extend subscriber activity ([84cd6a5](https://github.com/misikdmytro/gotick/commit/84cd6a575dc9a3adbfb5501878c562d138d6e7b5))
* start implementing scheduler [skip ci] ([64eb2fb](https://github.com/misikdmytro/gotick/commit/64eb2fb7990ad808855f748eb68eeb3122c3da58))
* test CD ([d1e566e](https://github.com/misikdmytro/gotick/commit/d1e566e87ed636983f5c382e38684cd5d282d6e0))
* update driver & factory registration ([0ada207](https://github.com/misikdmytro/gotick/commit/0ada207eec625aa58cffb0e5d235a8b46fd0a9b0))
* update exported modules from the package ([083224f](https://github.com/misikdmytro/gotick/commit/083224f06f5a81e2780fc60196bd02bdf9b8af4d))
* update husky to allow commit message length to be 1000 symbols ([6ffc680](https://github.com/misikdmytro/gotick/commit/6ffc680194efe51d07d6124eed53351ba79c6028))
* update planner & scheduler & cover planner with tests ([6a2e303](https://github.com/misikdmytro/gotick/commit/6a2e30335ec449fd2a02757db6c0c5e60cea8b5c))
* update public API, update settings & implement integration tests ([c9a679f](https://github.com/misikdmytro/gotick/commit/c9a679ff47a30a92bb8ae01bea1b5f72e8af709f))
* update schedule interface [skip ci] ([1159453](https://github.com/misikdmytro/gotick/commit/115945383855660926e67c5ec5db0cfab9e62346))
* update scheduler interface & cover with more tests ([853cbbf](https://github.com/misikdmytro/gotick/commit/853cbbfd529f7b1e2e1b2f54030d7e63eace5c5a))
* update schedules ([0c507f9](https://github.com/misikdmytro/gotick/commit/0c507f9135cd9d039d47b1a8e4c2cf9446284247))

# [0.1.0-feature-15-inquiry-implement-integration-benchmark-tests.1](https://github.com/misikdmytro/gotick/compare/v0.0.1...v0.1.0-feature-15-inquiry-implement-integration-benchmark-tests.1) (2024-09-15)


### Bug Fixes

* bypass SchedulerID to job context & improve tests for this ([bd947d6](https://github.com/misikdmytro/gotick/commit/bd947d65985376b5ed9da804d3d9b93ccd49fe49))
* fix driver registration problem with returning pointer instead of value itself ([361a66c](https://github.com/misikdmytro/gotick/commit/361a66cfe6ffa2463c1d3f5425a913016636e528))
* fix nil pointer dereference ([97a10a1](https://github.com/misikdmytro/gotick/commit/97a10a14fe1b42c772392af23345368f389c0b0d))
* mark job as currently executed in driver after the event from scheduler ([690fc64](https://github.com/misikdmytro/gotick/commit/690fc6485c70b2862fa1e39125c87b01106b3f5e))
* race condition for in-memory driver ([81f9eaa](https://github.com/misikdmytro/gotick/commit/81f9eaaae9cffd8d9ef62fabf0dae590d2aec53a))


### Features

* allow multi-threaded scheduler ([c5025a1](https://github.com/misikdmytro/gotick/commit/c5025a1cc31cb8eb8f2015dd1d6f8e939a12ab26))
* implement in memory driver (without tests for now) [skip ci] ([13e01dc](https://github.com/misikdmytro/gotick/commit/13e01dcf6f5e54e3b4937563a5d643d689f73638))
* implement job planner [skip ci] ([d05bc17](https://github.com/misikdmytro/gotick/commit/d05bc17c6b2a3b2f68e88e438f06ccf56ac0d2ec))
* implement job scheduling policies ([2181887](https://github.com/misikdmytro/gotick/commit/21818876bb979556450bcae23a28c1492cbb43d4))
* refactor planner &scheduler + new tests ([243ad99](https://github.com/misikdmytro/gotick/commit/243ad99796a605346efb309e94f0228e694fdfa1))
* remove errors propagation from subscriber _. simplify code ([8ecc04a](https://github.com/misikdmytro/gotick/commit/8ecc04a8d8900ac04468e068d6c16d3386fbbc4a))
* simplify interface by removing propagating errors and extend subscriber activity ([84cd6a5](https://github.com/misikdmytro/gotick/commit/84cd6a575dc9a3adbfb5501878c562d138d6e7b5))
* start implementing scheduler [skip ci] ([64eb2fb](https://github.com/misikdmytro/gotick/commit/64eb2fb7990ad808855f748eb68eeb3122c3da58))
* test CD ([d1e566e](https://github.com/misikdmytro/gotick/commit/d1e566e87ed636983f5c382e38684cd5d282d6e0))
* update driver & factory registration ([0ada207](https://github.com/misikdmytro/gotick/commit/0ada207eec625aa58cffb0e5d235a8b46fd0a9b0))
* update exported modules from the package ([083224f](https://github.com/misikdmytro/gotick/commit/083224f06f5a81e2780fc60196bd02bdf9b8af4d))
* update husky to allow commit message length to be 1000 symbols ([6ffc680](https://github.com/misikdmytro/gotick/commit/6ffc680194efe51d07d6124eed53351ba79c6028))
* update planner & scheduler & cover planner with tests ([6a2e303](https://github.com/misikdmytro/gotick/commit/6a2e30335ec449fd2a02757db6c0c5e60cea8b5c))
* update public API, update settings & implement integration tests ([c9a679f](https://github.com/misikdmytro/gotick/commit/c9a679ff47a30a92bb8ae01bea1b5f72e8af709f))
* update schedule interface [skip ci] ([1159453](https://github.com/misikdmytro/gotick/commit/115945383855660926e67c5ec5db0cfab9e62346))
* update scheduler interface & cover with more tests ([853cbbf](https://github.com/misikdmytro/gotick/commit/853cbbfd529f7b1e2e1b2f54030d7e63eace5c5a))
* update schedules ([0c507f9](https://github.com/misikdmytro/gotick/commit/0c507f9135cd9d039d47b1a8e4c2cf9446284247))
