# [0.1.0-feature-in-memory-driver.6](https://github.com/misikdmytro/gotick/compare/v0.1.0-feature-in-memory-driver.5...v0.1.0-feature-in-memory-driver.6) (2024-09-13)


### Features

* update schedule interface [skip ci] ([1ba5c1b](https://github.com/misikdmytro/gotick/commit/1ba5c1bcfa0cf26321474629d5c423b330df47e8))
* update schedules ([e23d599](https://github.com/misikdmytro/gotick/commit/e23d5991da71bdb0536678c6e65050ecdd0f41b8))

# [0.1.0-feature-in-memory-driver.5](https://github.com/misikdmytro/gotick/compare/v0.1.0-feature-in-memory-driver.4...v0.1.0-feature-in-memory-driver.5) (2024-09-13)


### Features

* simplify interface by removing propagating errors and extend subscriber activity ([96852ef](https://github.com/misikdmytro/gotick/commit/96852ef4dc0149583b946b4bc498cd319e96f110))

# [0.1.0-feature-in-memory-driver.4](https://github.com/misikdmytro/gotick/compare/v0.1.0-feature-in-memory-driver.3...v0.1.0-feature-in-memory-driver.4) (2024-09-13)


### Bug Fixes

* bypass SchedulerID to job context & improve tests for this ([04c5733](https://github.com/misikdmytro/gotick/commit/04c5733bf2768e49fb6bd01175e192ba487135a0))

# [0.1.0-feature-in-memory-driver.3](https://github.com/misikdmytro/gotick/compare/v0.1.0-feature-in-memory-driver.2...v0.1.0-feature-in-memory-driver.3) (2024-09-13)


### Bug Fixes

* mark job as currently executed in driver after the event from scheduler ([28cb4d3](https://github.com/misikdmytro/gotick/commit/28cb4d39f9b85a4a384313c15fb4350cda57f071))

# [0.1.0-feature-in-memory-driver.2](https://github.com/misikdmytro/gotick/compare/v0.1.0-feature-in-memory-driver.1...v0.1.0-feature-in-memory-driver.2) (2024-09-13)


### Bug Fixes

* fix driver registration problem with returning pointer instead of value itself ([cba5c8f](https://github.com/misikdmytro/gotick/commit/cba5c8fbb0e5bdbf983431c0c7ae5cec4423f110))

# [0.1.0-feature-in-memory-driver.1](https://github.com/misikdmytro/gotick/compare/v0.0.1...v0.1.0-feature-in-memory-driver.1) (2024-09-13)


### Bug Fixes

* fix nil pointer dereference ([f537d45](https://github.com/misikdmytro/gotick/commit/f537d4509e3417f3fc2cba5bec2a489e53a2075c))


### Features

* implement in memory driver (without tests for now) [skip ci] ([4718231](https://github.com/misikdmytro/gotick/commit/47182314597e9f382d4342867cc1044d951562ac))
* implement job planner [skip ci] ([d05bc17](https://github.com/misikdmytro/gotick/commit/d05bc17c6b2a3b2f68e88e438f06ccf56ac0d2ec))
* implement job scheduling policies ([2181887](https://github.com/misikdmytro/gotick/commit/21818876bb979556450bcae23a28c1492cbb43d4))
* refactor planner &scheduler + new tests ([73dd4be](https://github.com/misikdmytro/gotick/commit/73dd4be84bfefa1570d2e3f720884f36fe499c82))
* remove errors propagation from subscriber _. simplify code ([92e13cb](https://github.com/misikdmytro/gotick/commit/92e13cbed2be0fe2a4286dcc6a5cf3b0a49b44d3))
* start implementing scheduler [skip ci] ([64eb2fb](https://github.com/misikdmytro/gotick/commit/64eb2fb7990ad808855f748eb68eeb3122c3da58))
* test CD ([d1e566e](https://github.com/misikdmytro/gotick/commit/d1e566e87ed636983f5c382e38684cd5d282d6e0))
* update exported modules from the package ([4056f6b](https://github.com/misikdmytro/gotick/commit/4056f6b11daeeb212b7d1c1b5cb178a0948124c8))
* update husky to allow commit message length to be 1000 symbols ([6ffc680](https://github.com/misikdmytro/gotick/commit/6ffc680194efe51d07d6124eed53351ba79c6028))
* update planner & scheduler & cover planner with tests ([c171d69](https://github.com/misikdmytro/gotick/commit/c171d69e58618992050548a006de35cc5b3a86ea))
* update scheduler interface & cover with more tests ([1b1b244](https://github.com/misikdmytro/gotick/commit/1b1b244529273ed48bd30d5e18e0baf768654d48))

# [0.1.0-feature-in-memory-driver.1](https://github.com/misikdmytro/gotick/compare/v0.0.1...v0.1.0-feature-in-memory-driver.1) (2024-09-12)


### Bug Fixes

* fix nil pointer dereference ([f537d45](https://github.com/misikdmytro/gotick/commit/f537d4509e3417f3fc2cba5bec2a489e53a2075c))


### Features

* implement in memory driver (without tests for now) [skip ci] ([4718231](https://github.com/misikdmytro/gotick/commit/47182314597e9f382d4342867cc1044d951562ac))
* implement job planner [skip ci] ([d05bc17](https://github.com/misikdmytro/gotick/commit/d05bc17c6b2a3b2f68e88e438f06ccf56ac0d2ec))
* implement job scheduling policies ([2181887](https://github.com/misikdmytro/gotick/commit/21818876bb979556450bcae23a28c1492cbb43d4))
* refactor planner &scheduler + new tests ([73dd4be](https://github.com/misikdmytro/gotick/commit/73dd4be84bfefa1570d2e3f720884f36fe499c82))
* remove errors propagation from subscriber _. simplify code ([92e13cb](https://github.com/misikdmytro/gotick/commit/92e13cbed2be0fe2a4286dcc6a5cf3b0a49b44d3))
* start implementing scheduler [skip ci] ([64eb2fb](https://github.com/misikdmytro/gotick/commit/64eb2fb7990ad808855f748eb68eeb3122c3da58))
* test CD ([d1e566e](https://github.com/misikdmytro/gotick/commit/d1e566e87ed636983f5c382e38684cd5d282d6e0))
* update exported modules from the package ([4056f6b](https://github.com/misikdmytro/gotick/commit/4056f6b11daeeb212b7d1c1b5cb178a0948124c8))
* update husky to allow commit message length to be 1000 symbols ([6ffc680](https://github.com/misikdmytro/gotick/commit/6ffc680194efe51d07d6124eed53351ba79c6028))
* update planner & scheduler & cover planner with tests ([c171d69](https://github.com/misikdmytro/gotick/commit/c171d69e58618992050548a006de35cc5b3a86ea))
* update scheduler interface & cover with more tests ([1b1b244](https://github.com/misikdmytro/gotick/commit/1b1b244529273ed48bd30d5e18e0baf768654d48))

# [0.1.0-beta.1](https://github.com/misikdmytro/gotick/compare/v0.0.1...v0.1.0-beta.1) (2024-09-12)


### Bug Fixes

* fix nil pointer dereference ([9a69d1a](https://github.com/misikdmytro/gotick/commit/9a69d1a86d34365fdf8d3577252ff830b92ed221))


### Features

* implement job planner [skip ci] ([4a6ffba](https://github.com/misikdmytro/gotick/commit/4a6ffbae8993b9e572a8d37da5b99fe0558b60af))
* implement job scheduling policies ([6a16e9d](https://github.com/misikdmytro/gotick/commit/6a16e9db8b07377e9b0c6fe5b2d934985ffe4ef9))
* refactor planner &scheduler + new tests ([a956cc1](https://github.com/misikdmytro/gotick/commit/a956cc1f9f0ba1871133e072926844c026c0c8f6))
* remove errors propagation from subscriber _. simplify code ([d7797f5](https://github.com/misikdmytro/gotick/commit/d7797f5bdc8398f2ededbf79dcfccfa1a248527b))
* start implementing scheduler [skip ci] ([51215ed](https://github.com/misikdmytro/gotick/commit/51215edf7dc5202f0ce54f00a19d02d05eec41ad))
* test CD ([d1e566e](https://github.com/misikdmytro/gotick/commit/d1e566e87ed636983f5c382e38684cd5d282d6e0))
* update exported modules from the package ([087e1b1](https://github.com/misikdmytro/gotick/commit/087e1b14ca91521a06f82560264d65adfa9d0f5c))
* update husky to allow commit message length to be 1000 symbols ([6ffc680](https://github.com/misikdmytro/gotick/commit/6ffc680194efe51d07d6124eed53351ba79c6028))
* update planner & scheduler & cover planner with tests ([01c998c](https://github.com/misikdmytro/gotick/commit/01c998c835dab8127741062a4fdd1ef7243fc602))
* update scheduler interface & cover with more tests ([36a858d](https://github.com/misikdmytro/gotick/commit/36a858d66fdbe3beeb56fb454e486ba650766e8d))
