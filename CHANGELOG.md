<a name="EdgeX Record and Replay Changelog"></a>
## EdgeX Record and Replay

## Change Logs for EdgeX Dependencies

- [app-functions-sdk-go](https://github.com/edgexfoundry/app-functions-sdk-go/blob/main/CHANGELOG.md)

## [4.0.0] Odessa - 2025-03-12 (Only compatible with the 4.x releases)

### ✨  Features

- Enable PIE support for ASLR and full RELRO ([c510f15…](https://github.com/edgexfoundry/app-record-replay/commit/c510f15927f62f99bdd37ed4cd52a8e0d034abc9))

### ♻ Code Refactoring

- Update module to v4 ([a65e098…](https://github.com/edgexfoundry/app-record-replay/commit/a65e098de420ba3dc87bc70fddc9b356c65c4cb8))
```text

BREAKING CHANGE: import paths will need to change to v4, also replace consul with keeper

```

### 🐛 Bug Fixes

- Only one ldflags flag is allowed ([f726f1f…](https://github.com/edgexfoundry/app-record-replay/commit/f726f1f1e54e453d465af1427730aa5e1b37ac9c))

### 📖 Documentation

- Move API document files from openapi/v3 to openapi ([0358623…](https://github.com/edgexfoundry/app-rfid-llrp-inventory/commit/03586239b5f27a5097d49cb846f66fc97f92ef04))

### 👷 Build

- Upgrade to go-1.23, Linter1.61.0 and Alpine 3.20 ([3b3a4b3…](https://github.com/edgexfoundry/app-record-replay/commit/3b3a4b358737e2da1c426ed6ab247ca110a4e992))

## [3.1.0] Napa - 2023-11-15 (Only compatible with the 3.x releases)

This service is new in the Napa 3.1 release. 
Please refer to [the documentation] (https://docs.edgexfoundry.org/3.1/microservices/application/services/AppRecordReplay/Purpose/) for more details

