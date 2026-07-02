# Changelog

All notable changes to this project are documented here, newest first.

Entries are generated from [Conventional Commits](https://www.conventionalcommits.org)
and grouped by change type. This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Features

- Accumulate repeated slice flags instead of overwriting by [@iberflow](https://github.com/iberflow) in [b6c5327](https://github.com/toaweme/cli/commit/b6c53279155524352b576afa72d8c115b8cc529b).

### Chores & Other

- Bump structs to v0.4.0 by [@iberflow](https://github.com/iberflow) in [5fe2274](https://github.com/toaweme/cli/commit/5fe22740932cc5a2e614156501b10fe259146565).
- Relicense from MIT to Apache 2.0 by [@iberflow](https://github.com/iberflow) in [130d076](https://github.com/toaweme/cli/commit/130d0766f02adef34cd343949e9d43d38ca9eb46).

## [0.3.3] - 2026-07-01

### Chores & Other

- Bump toaweme deps to latest releases by [@iberflow](https://github.com/iberflow) in [5b97c6c](https://github.com/toaweme/cli/commit/5b97c6c6b08a37dbc4cc4e6cd7bf336422f45526).

## [0.3.2] - 2026-07-01

### Fixes

- Replace deprecated reflect.Ptr with reflect.Pointer by [@iberflow](https://github.com/iberflow) in [2615b78](https://github.com/toaweme/cli/commit/2615b780695571552af0986f46add86c18402ad8).
- Ci publish mend report once by [@iberflow](https://github.com/iberflow) in [b130c73](https://github.com/toaweme/cli/commit/b130c73ef1ca5703118d7699581312f180ecfc97).
- Ci workflow by [@iberflow](https://github.com/iberflow) in [5db9001](https://github.com/toaweme/cli/commit/5db9001b6f4dd9aba1d4b5c74b9a28b17e9971f0).

### CI & Build

- Bump care to v0.8.0 by [@iberflow](https://github.com/iberflow) in [de01ef8](https://github.com/toaweme/cli/commit/de01ef880baa1c5105cd9282084d4afebc69352a).
- Bump care to v0.7.1 and pin to commit sha by [@iberflow](https://github.com/iberflow) in [d00cc1d](https://github.com/toaweme/cli/commit/d00cc1d72931dd59dfe6649b52aa90e0bec27911).
- Bump care to v0.6.0 and fix card-svg dark/light wiring by [@iberflow](https://github.com/iberflow) in [62929ac](https://github.com/toaweme/cli/commit/62929acf40a0a6bf10e4876b94d997d0b1cb08bd).

### Chores & Other

- Pin ci care version and bump deps by [@iberflow](https://github.com/iberflow) in [0b21357](https://github.com/toaweme/cli/commit/0b2135791cfe1d8a63b55378c40fa95402c3a71b).
- Align README, CHANGELOG, and quality workflow with org standards by [@iberflow](https://github.com/iberflow) in [447c71f](https://github.com/toaweme/cli/commit/447c71f4b2ea943e54b6ac05d94fdac805b2f6a6).

## [0.3.1] - 2026-06-21

### Fixes

- Drop verbosity struct arg tag and show it off in readme by [@iberflow](https://github.com/iberflow) in [d3e4daa](https://github.com/toaweme/cli/commit/d3e4daaf4909004a8928ed9d2e2f26f17a5bdc03).

## [0.3.0] - 2026-06-21

### Features

- Clearer help, default-command block, aligned short flags by [@iberflow](https://github.com/iberflow) in [ff7b467](https://github.com/toaweme/cli/commit/ff7b4673d8839f1c3b18697ae625f81bdbbe04dc).

## [0.2.1] - 2026-06-21

### Fixes

- Report unknown commands instead of running the default by [@iberflow](https://github.com/iberflow) in [9e89c5b](https://github.com/toaweme/cli/commit/9e89c5bf818b2f82282e156eb1df9ab20d3d96e1).

### Chores & Other

- Merge pull request #3 from toaweme/vhs-demo-fix1 by [@zolia](https://github.com/zolia) in [00f11ad](https://github.com/toaweme/cli/commit/00f11adfdd156b42a5d1ae195444cdfab77ac7d1).
- Vhs demo for cli features by Antanas Masevicius in [14f1600](https://github.com/toaweme/cli/commit/14f1600fbbdb88421925b6d8f36f15729c784230).
- Merge pull request #1 from toaweme/vhs-demo by [@zolia](https://github.com/zolia) in [036b600](https://github.com/toaweme/cli/commit/036b6003fbebbd620bb8034e6f54f87c3a0c0e7f).
- Vhs demo for cli features by Antanas Masevicius in [62004ff](https://github.com/toaweme/cli/commit/62004ffd3d352a153f233d1bde8e470932f7a64f).
- Update taskfile by [@iberflow](https://github.com/iberflow) in [5d22edc](https://github.com/toaweme/cli/commit/5d22edc3ac0a5f0d5b85a0205108bc218c8eb8bf).

## [0.2.0] - 2026-06-12

### Chores & Other

- Cleanup codecs + taskfile by [@iberflow](https://github.com/iberflow) in [2ac2025](https://github.com/toaweme/cli/commit/2ac2025e77a12560bb16a2d0b4836e8b19d62d3d).
- Update readme by [@iberflow](https://github.com/iberflow) in [ef7ad7b](https://github.com/toaweme/cli/commit/ef7ad7b5b3cc72fc5f8656712769fd78ecdddf89).

## [0.1.0] - 2026-06-12

### Features

- Golangci linter starting point by [@iberflow](https://github.com/iberflow) in [34106cc](https://github.com/toaweme/cli/commit/34106cc44ac2cfe9001f234c55833b8b89871bc7).
- Bump deps by [@iberflow](https://github.com/iberflow) in [0ad9efa](https://github.com/toaweme/cli/commit/0ad9efa39b707c9faa39ffb4e8364e7008745d2b).
- Pass unknowns by [@iberflow](https://github.com/iberflow) in [91b26b4](https://github.com/toaweme/cli/commit/91b26b405d7060f2fdbc81213c04a055a0a88388).
- Env vars by [@iberflow](https://github.com/iberflow) in [95e64f5](https://github.com/toaweme/cli/commit/95e64f5f16fa26d14af9e9e00106bdba1f505f55).
- Default command with env support by [@iberflow](https://github.com/iberflow) in [847a736](https://github.com/toaweme/cli/commit/847a7369f4e408c10dc61c4f1469c8b9a766e044).
- Demo command by [@iberflow](https://github.com/iberflow) in [9cc3ad4](https://github.com/toaweme/cli/commit/9cc3ad462433d7eebd3592e09645cd53ae87f8d1).
- Placeholder command to show sub commands by [@iberflow](https://github.com/iberflow) in [2b30b79](https://github.com/toaweme/cli/commit/2b30b79fae6dd9751aa9d1c15c1f807fb953a4b4).
- Version command by [@iberflow](https://github.com/iberflow) in [11c40d2](https://github.com/toaweme/cli/commit/11c40d26507b98ce55f818a06e5932e1a79fe531).
- --agent output modes by [@iberflow](https://github.com/iberflow) in [376597c](https://github.com/toaweme/cli/commit/376597cbb905b05d310c6e24098b7884ddd20900).
- Shell completion command by [@iberflow](https://github.com/iberflow) in [6e33d24](https://github.com/toaweme/cli/commit/6e33d24191ed4d94d6367e39be2c2010b4c103de).
- Help + dotenv by [@iberflow](https://github.com/iberflow) in [bdd4113](https://github.com/toaweme/cli/commit/bdd41130b3bb69f62c70cc6a91bde3acc6c4a26f).
- Built-in config by [@iberflow](https://github.com/iberflow) in [a61bcd9](https://github.com/toaweme/cli/commit/a61bcd9aca49bd79184847f59ba0d502ead348a7).
- Dev command by [@iberflow](https://github.com/iberflow) in [34278dd](https://github.com/toaweme/cli/commit/34278dd07430484fc188879f607a8f9a5d0a3572).
- Docs by [@iberflow](https://github.com/iberflow) in [dc887f1](https://github.com/toaweme/cli/commit/dc887f18e72e95679235f4879af160d4ec626610).
- Multiline command by [@iberflow](https://github.com/iberflow) in [b9135aa](https://github.com/toaweme/cli/commit/b9135aafc91416d8a14859f1021d823b6660b4da).
- Layered config merge into command configs with strategy and mapping by [@iberflow](https://github.com/iberflow) in [ebf6abf](https://github.com/toaweme/cli/commit/ebf6abf6e455f855f5c15499b6eda06ca6434270).
- Render oneof allowed values for nested struct sub-fields in help by [@iberflow](https://github.com/iberflow) in [d595e5f](https://github.com/toaweme/cli/commit/d595e5f368b0dde278dfb6f6bbd0ac064939ed84).
- Pluggable help output codecs via Config.Formats (yaml/toml --format) by [@iberflow](https://github.com/iberflow) in [ff1f1dc](https://github.com/toaweme/cli/commit/ff1f1dccfe800e28f71f91091c1629926135aaca).
- Full example with 3rd parties by [@iberflow](https://github.com/iberflow) in [4c06e37](https://github.com/toaweme/cli/commit/4c06e37d05332a16562b49c2e4bc3c00b9d9f255).
- **Help:** Add --help-values to show resolved flag values across all formats by [@iberflow](https://github.com/iberflow) in [66a1676](https://github.com/toaweme/cli/commit/66a16761a5cc67cd8bf27c160eef69660e201f44).
- --help-format rename sweep, flag-only --version, IsRealError helper by [@iberflow](https://github.com/iberflow) in [18d6fbe](https://github.com/toaweme/cli/commit/18d6fbed925895a6e257e2bbf0fb2ebb201fa180).
- Optional verbosity by [@iberflow](https://github.com/iberflow) in [f32aa82](https://github.com/toaweme/cli/commit/f32aa82ce6b27ac576a23c8e3cf0ba2ce8bc6679).
- MIT license by [@iberflow](https://github.com/iberflow) in [6f17cfc](https://github.com/toaweme/cli/commit/6f17cfc90b5c2a20d726249130e1ba826ce455f1).
- README + ci tests by [@iberflow](https://github.com/iberflow) in [b583ffa](https://github.com/toaweme/cli/commit/b583ffaf5c60739de10c09e6f486a0006f997396).

### Fixes

- Structs dep by [@iberflow](https://github.com/iberflow) in [586eeec](https://github.com/toaweme/cli/commit/586eeec1b7995714a0bd96e19302b282222f966b).
- Command search by [@iberflow](https://github.com/iberflow) in [a31ba1a](https://github.com/toaweme/cli/commit/a31ba1adf6c178e6332326669caad87aa1c98372).
- Args by [@iberflow](https://github.com/iberflow) in [f2385a3](https://github.com/toaweme/cli/commit/f2385a3c8b999a18a45ce04fb975c02a4c71e387).
- Arg processing by [@iberflow](https://github.com/iberflow) in [c61eb98](https://github.com/toaweme/cli/commit/c61eb98c9c1f73c30a16971d64f48affe9b2cfea).
- Argument parsing and help by [@iberflow](https://github.com/iberflow) in [ebf274b](https://github.com/toaweme/cli/commit/ebf274b7555610959abe84c18acc3ea98d86dcd7).
- Help when no args given by [@iberflow](https://github.com/iberflow) in [d9d8d0f](https://github.com/toaweme/cli/commit/d9d8d0fff038458828d2e765070ebcb059c7d6f0).
- Cli help by [@iberflow](https://github.com/iberflow) in [4004d6b](https://github.com/toaweme/cli/commit/4004d6b4b376821fb7dcbbafdfc9dd1a2876f0da).
- Set go.mod go version to 1.18 by [@iberflow](https://github.com/iberflow) in [05d9e7b](https://github.com/toaweme/cli/commit/05d9e7b4cd8d13bde3d351c42a0e68291f1f0ea6).
- Help displays sub-commands by [@iberflow](https://github.com/iberflow) in [928b499](https://github.com/toaweme/cli/commit/928b49972ca6386c9f5831ab985e88ae92557e77).
- Run command validations by [@iberflow](https://github.com/iberflow) in [9431671](https://github.com/toaweme/cli/commit/94316716365a3bfd519c14af93b73ee6a0ae974f).
- Remove global option shorthand names by [@iberflow](https://github.com/iberflow) in [2eeae18](https://github.com/toaweme/cli/commit/2eeae185542cbb0bae8566d5456e944ed390f85e).
- Help subcommand formatting by [@iberflow](https://github.com/iberflow) in [62421a8](https://github.com/toaweme/cli/commit/62421a851c08a7c89b921c7e9f8a5bebc6b58d1c).
- Interface by [@iberflow](https://github.com/iberflow) in [dacf14f](https://github.com/toaweme/cli/commit/dacf14fb2a8c07916cafa365be9237bf0f6623f0).
- Run default command with CLI flags; split app.go into app_*.go by [@iberflow](https://github.com/iberflow) in [881b743](https://github.com/toaweme/cli/commit/881b743b119d98de5e38620c6966c11b857bbb78).
- Help output by [@iberflow](https://github.com/iberflow) in [4f64f30](https://github.com/toaweme/cli/commit/4f64f30e1aa81c1c61b87d3eb686400fa037eaad).
- Saner global flag names and shorts by [@iberflow](https://github.com/iberflow) in [a5db90c](https://github.com/toaweme/cli/commit/a5db90c5ab4cd64fd77c3cad9e1058303116be67).

### Documentation

- Add package doc comment and runnable godoc examples by [@iberflow](https://github.com/iberflow) in [25c1b12](https://github.com/toaweme/cli/commit/25c1b125c4d979361427aa91836a187380766ec6).

### Refactors

- Parsing by [@iberflow](https://github.com/iberflow) in [51e495d](https://github.com/toaweme/cli/commit/51e495ded1bed0419c3a9aabafc0870c0372ec85).
- Tidy up by [@iberflow](https://github.com/iberflow) in [705ac3d](https://github.com/toaweme/cli/commit/705ac3d3e36b952b1a317ac86f459d6de028c3cb).
- NewApp returns App, io.Writer seam for help renderers, drop dead exports by [@iberflow](https://github.com/iberflow) in [db272da](https://github.com/toaweme/cli/commit/db272dab4503f76f66243dfae979dde87b204042).
- Slim Config to serializable DTO, move Store/Formats to App setters, rename GlobalOptions -> GlobalFlags by [@iberflow](https://github.com/iberflow) in [9b2cb35](https://github.com/toaweme/cli/commit/9b2cb35fb442ac2aa2b66e80df661849d10ef542).
- Cli.Resolver + standalone config package + tidy up by [@iberflow](https://github.com/iberflow) in [b07568f](https://github.com/toaweme/cli/commit/b07568f75aa582fc4660af6d6b32f9b27d833d10).
- **Config:** Rename Scope to handler, consolidate interfaces into types.go by [@iberflow](https://github.com/iberflow) in [12fd4f6](https://github.com/toaweme/cli/commit/12fd4f68045ba19e15c372c0ba843067d1347ba6).
- Simplify config module by [@iberflow](https://github.com/iberflow) in [0f646fc](https://github.com/toaweme/cli/commit/0f646fceb8fb2a025fb07a5e187aaa8132cd6539).

### Tests

- Add coverage for commands and help rendering by [@iberflow](https://github.com/iberflow) in [d083dab](https://github.com/toaweme/cli/commit/d083dab6f81e2ed72cf06ecd5daa79529c8944bc).

### Chores & Other

- Initial commit :) by [@iberflow](https://github.com/iberflow) in [0eaa2f7](https://github.com/toaweme/cli/commit/0eaa2f70f1ad60b37470df58cb54747058f4df9a).
- Bump structs by [@iberflow](https://github.com/iberflow) in [4e21930](https://github.com/toaweme/cli/commit/4e219303e300876e930cb15906ad44ccbd1d3f26).
- Cleanup by [@iberflow](https://github.com/iberflow) in [8bafc03](https://github.com/toaweme/cli/commit/8bafc03a847a4c54e7d3a510fba506522227edc9).
- Bump structs by [@iberflow](https://github.com/iberflow) in [6c4f00b](https://github.com/toaweme/cli/commit/6c4f00b274084772446ae9c3fb6dad10c7b63513).
- Move to awee-ai org by [@iberflow](https://github.com/iberflow) in [6dcab12](https://github.com/toaweme/cli/commit/6dcab1292b81822fa8e37c8a4226f63ae470b103).
- Bump deps by [@iberflow](https://github.com/iberflow) in [d8423e3](https://github.com/toaweme/cli/commit/d8423e36db0abb420e89d02754180792f3cfc8f7).
- Bump deps by [@iberflow](https://github.com/iberflow) in [de648f4](https://github.com/toaweme/cli/commit/de648f4c1231c16b0964313af3492ed25863a7f6).
- Bump deps by [@iberflow](https://github.com/iberflow) in [5931f83](https://github.com/toaweme/cli/commit/5931f83ce1e99c806fb859465de932f45c070289).
- Move org by [@iberflow](https://github.com/iberflow) in [78019c4](https://github.com/toaweme/cli/commit/78019c400edd257e422815cd96505e9a320b4225).
- Tidy up by [@iberflow](https://github.com/iberflow) in [c4bbf7b](https://github.com/toaweme/cli/commit/c4bbf7bf9c3a3b299a5f63adb60f4289b86101bd).
- Bump structs module by [@iberflow](https://github.com/iberflow) in [7a643dc](https://github.com/toaweme/cli/commit/7a643dc690802f02ddf7738060023cc9d83854fe).
- Bump structs by [@iberflow](https://github.com/iberflow) in [35858e2](https://github.com/toaweme/cli/commit/35858e257fc187a29389e4e212f860561046e60c).
- Modernize by [@iberflow](https://github.com/iberflow) in [c6f199c](https://github.com/toaweme/cli/commit/c6f199ceea02f13db2777f2645333441943f0ec7).
- Ditch logging by [@iberflow](https://github.com/iberflow) in [fef578b](https://github.com/toaweme/cli/commit/fef578bb25a814ca56b529aa14e4e422d16bcff3).
- Examples by [@iberflow](https://github.com/iberflow) in [085d7b3](https://github.com/toaweme/cli/commit/085d7b34e2a8bdddba0cc715dcad1512358abbf2).
- Tidy up by [@iberflow](https://github.com/iberflow) in [c94c29e](https://github.com/toaweme/cli/commit/c94c29e8a2508d06d68c0d690354cff401295963).
- Ditch testify by [@iberflow](https://github.com/iberflow) in [c14f131](https://github.com/toaweme/cli/commit/c14f13182c7101dc1cb8943be15d34a06e2b1d30).
- Cleanup by [@iberflow](https://github.com/iberflow) in [fc02c24](https://github.com/toaweme/cli/commit/fc02c241311ff3804638daadf321ed487d2166bb).
- Better examples by [@iberflow](https://github.com/iberflow) in [84df66f](https://github.com/toaweme/cli/commit/84df66f2fd547dc0c539e19a0ca72b57c9c0dee6).
- Tidy up by [@iberflow](https://github.com/iberflow) in [b1b5d1a](https://github.com/toaweme/cli/commit/b1b5d1a907d3f0fabc7515785354fe540b0e6c24).
- Tidy up by [@iberflow](https://github.com/iberflow) in [3b2c08a](https://github.com/toaweme/cli/commit/3b2c08a9257d7a5d9912b38be8658736a40028fc).
- Improve help by [@iberflow](https://github.com/iberflow) in [482181c](https://github.com/toaweme/cli/commit/482181c2c31d5366c91380a027551a9034d5f4a2).
- Cleanup by [@iberflow](https://github.com/iberflow) in [1988eb7](https://github.com/toaweme/cli/commit/1988eb781ec87aecee0d5c7b492b7280b09ba0d4).
- Fix tests by [@iberflow](https://github.com/iberflow) in [b7fec56](https://github.com/toaweme/cli/commit/b7fec560a820ed0ed21ea015b2b3c94a7f183e77).
- Show arg/flag values by [@iberflow](https://github.com/iberflow) in [8d076cd](https://github.com/toaweme/cli/commit/8d076cd99c28a0d454e2701e1daa96c82aea6fcc).
- Bump structs + merge app files by [@iberflow](https://github.com/iberflow) in [fdf0227](https://github.com/toaweme/cli/commit/fdf022727f3e81c4ffc430f733c881f8a8eb90bb).
- Use consts for app strings by [@iberflow](https://github.com/iberflow) in [ab80c37](https://github.com/toaweme/cli/commit/ab80c375e8e0bd4dd3a5516d27f2706009ff83b1).
- Help cleanup by [@iberflow](https://github.com/iberflow) in [c611aa3](https://github.com/toaweme/cli/commit/c611aa3bd9e38700183c8eea4ab650d830dc42a4).
- Cleanup by [@iberflow](https://github.com/iberflow) in [edcfe38](https://github.com/toaweme/cli/commit/edcfe380bc96c304ee1428e9cafe53badd081d5d).
- Add bins to .gitignore by [@iberflow](https://github.com/iberflow) in [7a8cccd](https://github.com/toaweme/cli/commit/7a8cccdc05cc5aa1a23f0be9bb737fc22c9bed09).
- Docs by [@iberflow](https://github.com/iberflow) in [5a9b6f5](https://github.com/toaweme/cli/commit/5a9b6f5ea12747b42a33fd3daea42bfc72e20aa7).
- Bump structs to 0.1.0 by [@iberflow](https://github.com/iberflow) in [f594ea0](https://github.com/toaweme/cli/commit/f594ea02d7bc2a0daa2542d7717115d84f4ee369).
- Cleanup by [@iberflow](https://github.com/iberflow) in [67c376a](https://github.com/toaweme/cli/commit/67c376a0a76d754a26a2b11faff4f6fb97e0c329).
- Lint cleanup, explicit config absence errors, structs v0.2.0 by [@iberflow](https://github.com/iberflow) in [2269d0a](https://github.com/toaweme/cli/commit/2269d0a5bb5fb3f235cd89919875efcc03eb533d).

## [config/addons/yaml/v0.2.0] - 2026-06-12

### Chores & Other

- Cleanup codecs + taskfile by [@iberflow](https://github.com/iberflow) in [2ac2025](https://github.com/toaweme/cli/commit/2ac2025e77a12560bb16a2d0b4836e8b19d62d3d).
- Update readme by [@iberflow](https://github.com/iberflow) in [ef7ad7b](https://github.com/toaweme/cli/commit/ef7ad7b5b3cc72fc5f8656712769fd78ecdddf89).

## [config/addons/toml/v0.2.0] - 2026-06-12

### Chores & Other

- Cleanup codecs + taskfile by [@iberflow](https://github.com/iberflow) in [2ac2025](https://github.com/toaweme/cli/commit/2ac2025e77a12560bb16a2d0b4836e8b19d62d3d).
- Update readme by [@iberflow](https://github.com/iberflow) in [ef7ad7b](https://github.com/toaweme/cli/commit/ef7ad7b5b3cc72fc5f8656712769fd78ecdddf89).

[Unreleased]: https://github.com/toaweme/cli/compare/v0.3.3...HEAD
[0.3.3]: https://github.com/toaweme/cli/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/toaweme/cli/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/toaweme/cli/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/toaweme/cli/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/toaweme/cli/compare/config/addons/yaml/v0.2.0...v0.2.1
[0.2.0]: https://github.com/toaweme/cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/toaweme/cli/releases/tag/v0.1.0
[config/addons/yaml/v0.2.0]: https://github.com/toaweme/cli/compare/v0.1.0...config/addons/yaml/v0.2.0
[config/addons/toml/v0.2.0]: https://github.com/toaweme/cli/compare/v0.1.0...config/addons/toml/v0.2.0
