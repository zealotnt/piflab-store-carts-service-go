# PIFLab Store API
[![CircleCI](https://circleci.com/gh/zealotnt/piflab-store-api-go.svg?style=svg)](https://circleci.com/gh/zealotnt/piflab-store-api-go)  
[![Coverage Status](https://coveralls.io/repos/github/zealotnt/piflab-store-api-go/badge.svg)](https://coveralls.io/github/zealotnt/piflab-store-api-go)  

## API Docs
http://docs.piflabstore.apiary.io/

## Dependencies
- **GO 1.5**

## 3rd parties

## Framework
- **Dependency**: [Godep](https://github.com/tools/godep)
- **Router**: [Gorilla Mux](https://github.com/gorilla/mux)

## Add package
- `go get <package>`
- `import "<package>"`
- `godep save ./...`

## Testing
`./testcoverage.sh`

## Migration

### Migrate
`goose up`

### Rollback
`goose down`

## Cart services structure
[Cart service structure document](./docs/cart-services-structure.md)
